class Logger:
    def __init__(self, name: str):
        self.name = name
        self.file = open("./logger.log", "a")

    def log(self, msg: str, *args):
        msg = "[%s] %s\n" % (self.name, msg)
        log_msg = "[%s] " % self.name
        log_msg += msg + '\n'
        log_msg = msg % args
        self.file.write(log_msg)

    def close(self):
        self.file.close()
