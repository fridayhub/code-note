#!-*- coding:utf-8 -*-
# !/usr/bin/python3

import requests
from lxml import etree
from urllib.parse import urlparse
import urllib.parse
import os
import time

base_url = 'http://81.110.88.39'
start_url = 'http://81.110.88.39/L%3A'
base_dir = '/home/lost+found'


def getHtml(url, i=0):
    print("getHtml", url)
    time.sleep(1.5)
    try:
        kv = {
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36'}
        r = requests.get(url, headers=kv, timeout=30)
        r.encoding = 'utf-8'
        return r.text

    except Exception as e:
        # code / reason / headers 异常
        print('get html httperror:{}'.format(e))
        if i < 5:
            getHtml(url, i+1)
        else:
            return ''


def getDocList(html):
    try:
        data = etree.HTML(html)
        tmp_urls = data.xpath('//td[@class="folder"]/span/nobr/a/@href')
        tmp_files = data.xpath('//td[@class="file"]/span/nobr/a/@href')
        return tmp_urls, tmp_files
    except Exception:
        print('lxml parse failed')
        return None, None


def createDir(dir):
    if not os.path.exists(dir):
        print('create:', dir)
        os.makedirs(dir)


def downFile(url):
    time.sleep(1)
    try:
        print('downloading:', url)
        turl = urllib.parse.unquote(url)
        path = base_dir + urlparse(turl).path
        createDir(os.path.split(path)[0])
        startTime = time.time()
        with requests.get(url, stream=True, timeout=5) as r:
            contentLength = int(r.headers['content-length'])
            line = 'content-length: %dB/ %.2fKB/ %.2fMB'
            line = line % (contentLength, contentLength / 1024, contentLength / 1024 / 1024)
            print(line)
            if checkLocalFile(path, contentLength):
                print("Already download")
                return 
            downSize = 0
            with open(path, 'wb') as f:
                for chunk in r.iter_content(8192):
                    if chunk:
                        f.write(chunk)
                    downSize += len(chunk)
                    line = '%d KB/s - %.2f MB， 共 %.2f MB'
                    line = line % (
                    downSize / 1024 / (time.time() - startTime), downSize / 1024 / 1024, contentLength / 1024 / 1024)
                    print(line, end='\r')
                    if downSize >= contentLength:
                        break
            timeCost = time.time() - startTime
            line = '共耗时: %.2f s, 平均速度: %.2f KB/s'
            line = line % (timeCost, downSize / 1024 / timeCost)
            print(line)
    except Exception as e:
        print(e)
        time.sleep(1)

def checkLocalFile(filePath, lenght):
    try:
        print(filePath)
        fileSize = os.path.getsize(filePath)
        return fileSize == lenght
    except Exception as e:
        return False

def down(urls):
    notDown = ['.cbr', '.cbz', '.jpg', '.png']
    if urls:
        for url in urls:
            try:
                turl = base_url + url
                print(turl)
                html = getHtml(turl)
                turls, tfiles = getDocList(html)
                if turls:
                    down(turls)
                for dfile in tfiles:
                   time.sleep(0.01)
                   if os.path.splitext(dfile)[-1].lower() in notDown or 'Elfquest' in dfile:
                       print("not down", dfile)
                       break
                   print('you need to down', base_url + dfile)
                   downFile(base_url + dfile)
            except Exception as e:
                print(e)
    else:
        print("user is null")


if __name__ == '__main__':
    html = getHtml(start_url)
    turls, tfiles = getDocList(html)
    down(turls)
    for file in tfiles:
        print('you need to down', base_url + file)
        downFile(base_url + file)
    # while True:
    #     html = getHtml(next_url)
    #     getDocList(html)
