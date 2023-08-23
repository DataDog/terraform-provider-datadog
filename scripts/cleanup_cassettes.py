import os
import re
import distutils

testNameRe = re.compile(r"func (Test\S*)\(t \*testing\.T\) {")

datadogTestsDir = "datadog/tests/"
datadogCassettesDir = "datadog/tests/cassettes/"


def extractTestNames(filename):
    testNames = set()
    with open(filename, mode="r") as f:
        for i, line in enumerate(f):
            match = testNameRe.match(line)
            if match:
                testNames.add(match.group(1))
    return testNames


def main():
    real_tests = set()
    for path, _, files in os.walk(datadogTestsDir):
        for name in files:
            if name.endswith("_test.go"):
                real_tests.update(extractTestNames(os.path.join(path, name)))

    files = [f for f in os.listdir(datadogCassettesDir) if os.path.isfile(
        os.path.join(datadogCassettesDir, f)) and f.split(".")[0] not in real_tests]
    for f in files:
        print(f)

    delete = distutils.util.strtobool(
        input("Delete the listed files? (y/n): "))
    if delete:
        for f in files:
            os.remove(os.path.join(datadogCassettesDir, f))


if __name__ == "__main__":
    main()
