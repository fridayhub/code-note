package rxdatabase

import (
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type Kvmap map[string]sql.RawBytes

type DbOper struct {
	Db      *sql.DB
	ErrMsg  string
	ErrCode uint16
}

// string convert to int32
func STI32(in string) int32 {
	out, err := strconv.Atoi(in)
	if err != nil {
		return 0
	}
	return int32(out)
}

func (d *DbOper) ParseError(err error) error {
	sqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		log.Debug(sqlErr.Message)
		log.Debug(sqlErr.Number)
		d.ErrCode = sqlErr.Number
		d.ErrMsg = sqlErr.Message
		return nil
	} else {
		return err
	}
}

func (d *DbOper) BeginTx() (*sql.Tx, error) {
	log.Debug("BeginTx")
	err := d.connRetires()
	if err != nil {
		return nil, err
	}
	return d.Db.Begin()
}

func (d *DbOper) RunTransSql(tx *sql.Tx, sqlss string) (int64, error) {
	log.Debug("RunTransSql:", sqlss)
	err := d.connRetires()
	if err != nil {
		return 0, err
	}
	ret, err := tx.Exec(sqlss)
	if err != nil {
		return -1, err
	}

	n, err := ret.RowsAffected()
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (d *DbOper) TransQueryRows(tx *sql.Tx, sqlss string, outType interface{}) ([]byte, error) {
	log.Debug("TransQueryRows:", sqlss)
	err := d.connRetires()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(sqlss)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	resultChan := make(chan []byte)
	kvMapChan := make(chan Kvmap)
	defer close(resultChan)

	go d.convertSqlRawBytes(kvMapChan, resultChan, outType)

	err = d.RowTokv(rows, kvMapChan)
	if err != nil {
		return nil, err
	}

	return <-resultChan, nil
}

func (d *DbOper) TransQueryRow(tx *sql.Tx, sqlss string, ret ...interface{}) error {
	log.Debug("GetRow:", sqlss)
	err := d.connRetires()
	if err != nil {
		return err
	}
	row := tx.QueryRow(sqlss)
	return row.Scan(ret...)
}

func (d *DbOper) CommitTx(tx *sql.Tx) error {
	log.Debug("CommitTx")
	err := d.connRetires()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DbOper) RollbackTx(tx *sql.Tx) error {
	log.Debug("RollbackTx")
	return tx.Rollback()
}

func (d *DbOper) GetLastInsertId(tx *sql.Tx) int32 {
	log.Debug("GetLastInsertId start...")
	sqlss := "SELECT LAST_INSERT_ID();"
	var id int32
	if tx != nil {
		_ = d.TransQueryRow(tx, sqlss, &id)
	} else {
		_ = d.GetRow(sqlss, &id)
	}
	log.Debug("GetLastInsertId end, id:", id)
	return id
}

func (d *DbOper) RunSql(sqlss string, args ...interface{}) (int64, error) {
	log.Debug("RunSql:", sqlss)
	err := d.connRetires()
	if err != nil {
		return 0, err
	}
	ret, err := d.Db.Exec(sqlss, args...)
	if err != nil {
		return -1, err
	}

	n, err := ret.RowsAffected()
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (d *DbOper) GetRow(sqlss string, ret ...interface{}) error {
	log.Debug("GetRow:", sqlss)
	err := d.connRetires()
	if err != nil {
		return err
	}
	row := d.Db.QueryRow(sqlss)
	err = row.Scan(ret...)
	if err != nil {
		return err
	}
	return nil
}

func (d *DbOper) RowTokv(rows *sql.Rows, kvMapChan chan Kvmap) (err error) {
	var (
		columnLen int
	)
	defer close(kvMapChan)
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return
	}
	columnLen = len(columns)
	// Make a slice for the values
	values := make([]sql.RawBytes, columnLen)

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, columnLen)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return
		}
		var line Kvmap
		line = make(Kvmap, columnLen)
		for i, col := range values {
			//fmt.Println("vvvvvvvvvv:", string(col))
			line[columns[i]] = col
		}
		kvMapChan <- line
	}

	if err = rows.Err(); err != nil {
		return
	}
	//	fmt.Println("result:", result)

	return
}

func (d *DbOper) convertSqlRawBytes(kvMapChan chan Kvmap, resultChan chan []byte, outType interface{}) {
	var convert []map[string]interface{}

	t := reflect.TypeOf(outType)
	if t.Kind() != reflect.Struct {
		return
	}
	vOf := reflect.ValueOf(outType)
	//fmt.Println(vOf.Field(0).Type().Name())
	for element := range kvMapChan {
		var tmp = make(map[string]interface{})
		for k, v := range element {
			switch vOf.FieldByName(k).Type().String() {
			case "sql.NullInt64":
				tmp[k] = rxs.STI64(string(v))
			case "sql.NullInt32":
				tmp[k] = rxs.STI32(string(v))
			case "sql.NullString":
				tmp[k] = string(v)
			case "sql.NullFloat64":
				tmp[k] = rxs.STF64(string(v))
			case "sql.NullTime":
				t, _ := time.Parse(time.RFC3339, string(v))
				tmp[k] = t.Format("2006-01-02 15:04:05")
			default:
				tmp[k] = v
			}
		}
		convert = append(convert, tmp)
	}
	ret, err := json.Marshal(convert)
	if err != nil {
		log.Error("marshal error:", err)
		return
	}
	resultChan <- ret
	return
}

func (d *DbOper) GetMulitRows(sqlss string, outType interface{}) (ret []byte, err error) {
	log.Debug("GetMulitRows:", sqlss)
	err = d.connRetires()
	if err != nil {
		return nil, err
	}
	rows, err := d.Db.Query(sqlss)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	resultChan := make(chan []byte)
	kvMapChan := make(chan Kvmap)
	defer close(resultChan)

	go d.convertSqlRawBytes(kvMapChan, resultChan, outType)

	err = d.RowTokv(rows, kvMapChan)
	if err != nil {
		return nil, err
	}

	return <-resultChan, nil
}

const maxRetires = 5

func (d *DbOper) connRetires() (err error) {
	for i := 0; i < maxRetires; i++ {
		err = d.Db.Ping()
		log.Debug("connRetires:", i, ", err:", err)
		if err == nil {
			return
		}
	}
	return
}
