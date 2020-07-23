import subprocess
import re
import sys
from math import ceil
from threading import Thread, RLock
from time import sleep


class Timeout(Thread):
    def __init__(self, lock):
        super().__init__()
        self.running = True
        self.lock = lock

    def run(self):
        i = 0
        while self.amIAlive():
            sleep(1)
            i = i + 1
            if i == 3600:
                with self.lock:
                    self.running = False

    def amIAlive(self) -> bool:
        with self.lock:
            return self.running

class Balanced(Thread):
    def __init__(self, actualWidth):
        super().__init__()
        self.actualWidth = actualWidth
        self.result = None
        self.running = True

    def run(self):
        self.process = subprocess.Popen(["./libs/balancedLinux", "-width", str(actualWidth), "-graph", hgFile, "-det"], stdout=subprocess.PIPE)
        self.stdout = self.process.communicate()[0]
        self.running = False



def findWidth(hgFile):
    hg = open(hgFile, "r")
    cont = 0
    for _ in hg:
        cont += 1
    return cont


if __name__ == '__main__':
    hgFile = sys.argv[1]
    actualWidth = ceil(findWidth(hgFile)/2)
    while True:
        balanced = Balanced(actualWidth)
        lockTimeout = RLock()
        timeout = Timeout(lockTimeout)
        timeout.start()
        balanced.start()
        while timeout.amIAlive() and balanced.running:
            sleep(0.1)
        if not timeout.amIAlive():
            print(actualWidth + 1)
            balanced.process.kill()
            break
        with lockTimeout:
            timeout.running = False
        result = balanced.stdout
        foundWidth = 0
        correct = False
        for line in str(result).split("\\n"):
            m = re.search("Width:\\s+(.*)", line)
            if m:
                foundWidth = int(m.group(1))
            m = re.search("Correct:\\s+(.*)", line)
            if m:
                if m.group(1) == "true":
                    correct = True
        if correct:
            actualWidth = foundWidth - 1
        else:
            print(actualWidth + 1)
            break
