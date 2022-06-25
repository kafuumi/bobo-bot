# encoding=UTF-8
import json
import math
import os.path
import sys
import time

import matplotlib.pyplot as plt
import numpy as np
import requests
import seaborn as sns
from matplotlib import font_manager as fm
from requests_toolbelt import MultipartEncoder
from logger import Logger


# 数据聚合
def gather(data: np.ndarray, step, func) -> np.ndarray:
    row = len(data) // step
    data = data[:row * step]  # 舍弃多余的元素
    data = data.reshape(row, step)
    r = np.zeros(0)
    for i in data:
        r = np.append(r, func(i))
    return r


def handle_resp(resp: requests.Response):
    if resp.status_code == 200:
        resp_data = resp.json()
        if resp_data['code'] != 0:
            logger.log("resp error: %s", resp_data['message'])
            return None
        return resp_data['data']
    else:
        logger.log("网络错误")
        return None


# 发布动态
def post_dynamic(msg: str, pics: list):
    now = time.time()
    pics_temp = []
    for p in pics:
        temp = {
            "img_src": p['image_url'],
            "img_width": p['image_width'],
            "img_height": p['image_height'],
            "img_size": p['img_size']
        }
        pics_temp.append(temp)

    data = {
        "dyn_req": {
            "content": {
                "contents": [
                    {
                        "raw_text": msg,
                        "type": 1,
                        "biz_id": ""
                    }
                ]
            },
            "pics": pics_temp,
            "meta": {
                "app_meta": {
                    "from": "create.dynamic.web",
                    "mobi_app": "web"
                }
            },
            "scene": 2,  # 有图片为2，无图为1
            "attach_card": None,
            "upload_id": "%s_%d_%d" % (cookie['DedeUserID'], int(now), int((now - int(now)) * 10000))
        }
    }
    # 无图片
    if pics is None or len(pics) == 0:
        data['dyn_req']['scene'] = 1
        del data['dyn_req']['pics']

    headers['Content-Type'] = "application/json; charset=utf-8"
    url = "https://api.bilibili.com/x/dynamic/feed/create/dyn?csrf=%s" % cookie['bili_jct']
    resp = requests.post(url, data=json.dumps(data), cookies=cookie, headers=headers)
    body = handle_resp(resp)
    if body is None:
        return
    resp_data = handle_resp(resp)
    if resp_data is None:
        return
    # print(resp_data)
    logger.log("发布动态成功！link: https://t.bilibili.com/%s" % resp_data['dyn_id_str'])


# 上传图片
def upload_img(img_path):
    data = {
        "file_up": (os.path.basename(img_path), open(img_path, 'rb'), 'image/jpeg'),
        "biz": "new_dyn",
        "category": "daily",
        "csrf": cookie['bili_jct']
    }
    data = MultipartEncoder(fields=data)
    headers['Content-Type'] = data.content_type
    url = "https://api.bilibili.com/x/dynamic/feed/draw/upload_bfs"
    resp = requests.post(url, data=data, headers=headers, cookies=cookie)
    resp_data = handle_resp(resp)
    if resp_data is None:
        return None
    logger.log("上传图片成功：%s", resp_data)
    resp_data['img_size'] = os.path.getsize(img_path) / 1024
    return resp_data


# 绘制折线图
def draw_plot(x, y, x_label, y_label, title, save_path, is_bar: bool = False, is_fill: bool = True):
    _len = len(y)
    step = 0
    if _len > 24:
        step = math.ceil(_len / 24)
    sns.set()
    plt.figure(figsize=(16, 9), dpi=100)
    if is_bar:
        sns.barplot(x=x[:_len], y=y)
        plt.grid(visible=True)
    else:
        if is_fill:
            plt.fill_between(x[:_len], y, color='skyblue', alpha=0.4)
        sns.lineplot(x=x[:_len], y=y)
    if step != 0:
        plt.xticks(np.arange(0, _len + step, step), rotation=33)
    y_max = np.max(y) + 10
    y_min = np.min(y)
    if y_min < 50:
        y_min = 0
    step = 1
    if y_max - y_min > 15:
        step = math.ceil((y_max - y_min) / 15)
    labels = np.arange(y_min, int(y_max + step), step)
    plt.yticks(ticks=labels, labels=labels)
    plt.title(title, fontproperties=selected_font)
    plt.xlabel(x_label, fontproperties=selected_font)
    plt.ylabel(y_label, fontproperties=selected_font)
    plt.savefig(save_path)
    plt.clf()


# 处理数据并绘图
def draw(board, fans: np.ndarray):
    start = board['start']
    end = board['end']
    hot = np.array(board['hot'])
    delay = np.array(board['awl'])

    time_str = np.arange(start, end + 60, 60)
    remain = len(time_str) % 10

    # 填充数据
    if remain != 0:
        for i in range(10 - remain):
            time_str = np.append(time_str, time_str[-1] + 60)
    remain = len(hot) % 10
    if remain != 0:
        for i in range(10 - remain):
            hot = np.append(hot, 0)
    remain = len(delay) % 10
    if remain != 0:
        for i in range(10 - remain):
            delay = np.append(delay, delay[-1])

    # 每10个数据聚合，一个数据代表一分钟内的数据，所以10个聚合则是10分钟内的数据
    step = 10
    time_str = np.array([time.strftime("%H:%M", time.localtime(t)) for t in time_str[::step]])
    hot = gather(hot, step, np.sum)
    # 延迟数据会受某些极端值影响，这里同时绘制平均值图像和中位数图像
    # 这部分极端值产生的原因可能和发布评论的账号有关，而不是和评论区有关
    # 取十分钟内的平均数
    delay_mean = gather(delay, step, np.mean)
    # 取十分钟内的中位数
    delay_median = gather(delay, step, np.median)

    img_dir = "./report/img"
    if not os.path.exists(img_dir):
        os.makedirs(img_dir)
    time_range = "%s - %s" % (time.strftime("%m-%d", time.localtime(start)),
                              time.strftime("%m-%d", time.localtime(end)))
    draw_plot(time_str, hot, "时间", "评论数",
              "%s 十分钟内总评论数" % time_range, img_dir + "/hot.jpg", True)
    draw_plot(time_str, delay_mean, "时间", "平均延迟",
              "%s 十分钟内平均延迟（单位：秒）" % time_range, img_dir + "/delay_mean.jpg")
    draw_plot(time_str, delay_median, "时间", "延迟中位数",
              "%s 十分钟内延迟中位数（单位：秒）" % time_range, img_dir + "/delay_median.jpg")
    draw_plot(time_str, fans, "时间", "粉丝数",
              "%s 粉丝数变化" % time_range, img_dir + "/fans.jpg", is_fill=False)


def main():
    if len(sys.argv) <= 1:
        logger.log("缺少输入文件")
        sys.exit(1)
    file_name = sys.argv[1]
    with open(file_name, encoding='utf-8') as f:
        data = json.load(f)
    board = data['board']
    account = data['account']
    draw(board, np.array(account['fansCount']))
    start_all_count = board['startAllCount']
    start_count = board['startCount']
    end_all_count = board['endAllCount']
    end_count = board['endCount']
    count = board['count']

    start_follower = account['startFollowers']
    end_follower = account['endFollowers']
    people = board['people']

    start_time = board['start']
    end_time = board['end']
    hot = np.array(board['hot'])
    max_hot_time = 0
    max_hot = 0
    for i in range(len(hot)):
        if hot[i] > max_hot:
            max_hot = hot[i]
            max_hot_time = i
    max_hot_time = board['start'] + max_hot_time * 60

    max_uid, max_num = 0, 0
    people = np.array(list(people.items()), dtype=int)
    for i in range(len(people)):
        item = people[i]
        if item[1] > max_num:
            max_uid = item[0]
            max_num = item[1]
    msg = '【数据总结】%s-%s\n' \
          '【%s】粉丝数变化：%d => %d(%+d)\n' \
          '【%s】评论数变化：%d => %d(%+d)\n' \
          '不含楼中楼评论数：%d => %d(%+d)\n' \
          '%s 达到最高同接：%d条/分钟\n' \
          '发送评论人数：%d\n' \
          '单个账号最多发送评论：%d 条' % (
              time.strftime("%m月%d日", time.localtime(start_time)),
              time.strftime("%m月%d日", time.localtime(end_time)),
              account['name'], start_follower, end_follower, end_follower - start_follower,
              board['name'], start_all_count, end_all_count, end_all_count - start_all_count,
              start_count, end_count, end_count - start_count,
              time.strftime("%m-%d %H:%M", time.localtime(max_hot_time)), max_hot,
              len(people), max_num
          )
    logger.log(msg)
    logger.log("记录的评论数：%d" % count)
    logger.log("最佳人之初：uid:%d" % max_uid)
    if len(sys.argv) == 2:
        return
    # 发布动态
    images = []
    hot_img = upload_img("./report/img/hot.jpg")
    if hot_img is None:
        logger.log("上传图片：hot失败")
        return
    images.append(hot_img)
    fans_img = upload_img("./report/img/fans.jpg")
    if fans_img is None:
        logger.log("上传图片：fans失败")
        return
    images.append(fans_img)
    delay_mean_img = upload_img("./report/img/delay_mean.jpg")
    if delay_mean_img is None:
        logger.log("上传图片，delay_mean失败")
        return
    images.append(delay_mean_img)
    delay_median_img = upload_img("./report/img/delay_median.jpg")
    if delay_median_img is None:
        logger.log("上传图片：delay_median失败")
        return
    images.append(delay_median_img)
    post_dynamic(msg, images)


if __name__ == '__main__':
    logger = Logger("python")

    font_list = dict()
    font_name = None
    for font in fm.FontManager().ttflist:
        font_list[font.name] = font.fname
    if 'SimHei' in font_list:
        font_name = 'SimHei'
    elif 'SimSun' in font_list:
        font_name = 'SimSun'
    elif 'Microsoft YaHei' in font_list:
        font_name = 'Microsoft YaHei'
    if font_name is None:
        logger.log("未找到合适的字体,需要以下任一字体：宋体，黑体，微软雅黑")
        sys.exit(1)
    logger.log("使用字体：%s", font_name)
    selected_font = fm.FontProperties(fname=font_list[font_name])

    with open("./setting.json", encoding='utf-8') as setting:
        setting_json = json.load(setting)
        cookie = dict()
        cookie['DedeUserID'] = str(setting_json['botAccount']['uid'])
        cookie['DedeUserID__ckMd5'] = setting_json['botAccount']['uidMd5']
        cookie['SESSDATA'] = setting_json['botAccount']['sessData']
        cookie['bili_jct'] = setting_json['botAccount']['csrf']
        cookie['sid'] = setting_json['botAccount']['sid']
        headers = {
            'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) '
                          'Chrome/96.0.4664.93 Safari/537.36',
            'sec-ch-ua': 'ot A;Brand";v="99", "Chromium";v="96", "Google Chrome";v="96',
            'sec-ch-ua-mobile': '?0',
            'sec-ch-ua-platform': 'Windows',
            'accept-language': 'zh-CN,zh;q=0.9',
            "Accept-Encoding": "gzip, deflate, br"
        }
    main()
    logger.close()
