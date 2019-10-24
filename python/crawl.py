#!-*- coding:utf-8 -*-
# !/usr/bin/python3

import requests
from lxml import etree
from urllib.parse import urlparse
import urllib.parse
import os
import time

base_url = 'http://xxxxxx'
start_url = 'http://xxxxxxx'
base_dir = '/home/xxx/dxxxown'


def getHtml(url):
    print("getHtml", url)
    try:
        kv = {
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36'}
        r = requests.get(url, headers=kv, timeout=30)
        r.encoding = 'utf-8'
        return r.text


    except requests.URLError as e:
        print('get html urlerror:{}'.format(e))
        return ''

    except requests.HTTPError as e:
        # code / reason / headers 异常
        print('get html httperror:{}'.format(e))
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
    print('downloading:', url)
    turl = urllib.parse.unquote(url)
    path = base_dir + urlparse(turl).path
    createDir(os.path.split(path)[0])
    startTime = time.time()
    with requests.get(url, stream=True) as r:
        contentLength = int(r.headers['content-length'])
        line = 'content-length: %dB/ %.2fKB/ %.2fMB'
        line = line % (contentLength, contentLength / 1024, contentLength / 1024 / 1024)
        print(line)
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


def down(urls):
    if urls:
        for url in urls:
            try:
                turl = base_url + url
                print(turl)
                html = getHtml(turl)
                turls, tfiles = getDocList(html)
                if turls:
                    down(turls)
                for file in tfiles:
                    print('you need to down', base_url + file)
                    downFile(base_url + file)
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
