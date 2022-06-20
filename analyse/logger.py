# encoding=utf-8
import sys
import time


class Logger:
    def __init__(self, name: str):
        self.name = name
        # self.file = open("./logger.log", "a")

    def log(self, msg: str, *args):
        msg = "[%s] %s %s\n" % (self.name, time.strftime("%m-%d %H:%M"), msg)
        log_msg = msg % args
        sys.stdout.buffer.write(log_msg.encode(encoding='utf-8'))

    def close(self):
        # self.file.close()
        pass
