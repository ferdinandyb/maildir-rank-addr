import json
import re
from collections import Counter
from email.header import Header, decode_header
from email.parser import BytesParser
from email.utils import parsedate_to_datetime
from pathlib import Path

MAILDIR = "/home/fbence/.mail"
CACHEPATH = "addressbook"
FILTERLIST = ["do-not-reply", "no-reply", "bounce", "noreply"]
FREQUENCY_WEIGHT = 0.5

# RFC 5322 compliant regex from Moritz Poldrack (moritz@poldrack.dev)
ADDRREGEX = r"""(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])"""

PARSEADDRESS = re.compile(f"(.*?)<?({ADDRREGEX})>?")

COMMA_MATCHER = re.compile(r",(?=(?:[^\"']*[\"'][^\"']*[\"'])*[^\"']*$)")


def filterAddress(address):
    for filt in FILTERLIST:
        if filt in address.split("@")[0]:
            return True
    return False


def checkIfMail(path):
    if path.is_dir():
        return False
    if path.parent.stem in ["cur", "tmp", "new"]:
        return True
    return False


def parseAddress(address):
    parts = []
    for part, encoding in decode_header(address.strip()):
        try:
            if encoding is None and type(part) == bytes:
                part = part.decode()
            elif encoding is not None:
                part = part.decode(encoding)
            parts.append(part.strip().replace("\n", ""))
        except UnicodeDecodeError:
            return None
    address = " ".join(parts)
    m = PARSEADDRESS.search(address, re.IGNORECASE)
    if m:
        name, address = m.group(1, 2)
        name = name.replace('"', "").strip()
        address = address.lower().strip()
        return name, address
    else:
        return None


def parseHeader(header):
    if type(header) == Header or header is None or header.strip() == "":
        return None
    addresses = []
    for address in COMMA_MATCHER.split(header):
        address = parseAddress(address)
        if address is not None:
            addresses.append(address)
    return addresses


def parseMessage(msg):
    addresses = []
    for header in ["to", "from", "cc", "bcc"]:
        parsed = parseHeader(msg[header])
        if parsed is not None:
            addresses += parsed
    try:
        date = parsedate_to_datetime(msg["date"]).timestamp()
    except ValueError:
        date = 0
    return addresses, date


def getMostFrequent(List):
    occurence_count = Counter(List)
    return occurence_count.most_common(1)[0][0]


def loadCache(cachefilepath):
    if cachefilepath.exists():
        cache = json.loads(open(cachefilepath).read())
    else:
        cache = {"msgs_seen": [], "addresses": {}}
    parser = BytesParser()
    for i, msgpath in enumerate(Path(MAILDIR).glob("**/*")):
        if not checkIfMail(msgpath):
            continue
        msgpathid = msgpath.name.split(",")[0]
        if msgpathid in cache["msgs_seen"]:
            continue
        else:
            msg = parser.parse(open(msgpath, "rb"), headersonly=True)
            addresses, date = parseMessage(msg)
            for name, emailaddr in addresses:
                if not emailaddr in cache["addresses"]:
                    cache["addresses"][emailaddr] = {"names": [name], "dates": [date]}
                else:
                    cache["addresses"][emailaddr]["names"].append(name)
                    cache["addresses"][emailaddr]["dates"].append(date)

            cache["msgs_seen"].append(msgpathid)
    return cache


def calculateRanks(cache):
    frequency_ranks = []
    recency_ranks = []
    for key, value in cache["addresses"].items():
        frequency_ranks.append([key, len(value["dates"])])
        recency_ranks.append([key, max(value["dates"])])
    for i, (key, _) in enumerate(
        sorted(frequency_ranks, key=lambda x: x[1], reverse=True)
    ):
        cache["addresses"][key]["frequency_rank"] = i
    for i, (key, _) in enumerate(
        sorted(recency_ranks, key=lambda x: x[1], reverse=True)
    ):
        cache["addresses"][key]["recency_rank"] = i
    return cache


def getTotalRank(item):
    freq = item["frequency_rank"]
    rec = item["recency_rank"]
    return freq * FREQUENCY_WEIGHT + (1 - FREQUENCY_WEIGHT) * rec


def saveAercCompatibleOutput(cache, outputfile):
    total_ranks = []
    for addr, value in cache["addresses"].items():
        if filterAddress(addr):
            continue
        name = getMostFrequent(value["names"])
        if name is None or name == "":
            line = f"{addr}\n"
        else:
            line = f"{addr}\t{name}\n"
        total_ranks.append(
            {
                "line": line,
                "rank": getTotalRank(value),
            }
        )
    with open(outputfile, "w") as f:
        for item in sorted(total_ranks, key=lambda x: x["rank"]):
            f.write(item["line"])


def main():
    cachedir = Path(CACHEPATH)
    cachedir.mkdir(parents=True, exist_ok=True)
    cachefilepath = cachedir / "cache.json"
    outputfilepath = cachedir / "addresses.txt"
    cache = loadCache(cachefilepath)
    json.dump(cache, open(cachefilepath, "w"))
    cache = calculateRanks(cache)
    saveAercCompatibleOutput(cache, outputfilepath)


if __name__ == "__main__":
    main()
