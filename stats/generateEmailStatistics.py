from datetime import datetime
from email.parser import BytesParser
from pathlib import Path

from matplotlib import pyplot as plt

from generateAddressbookMaildir import MAILDIR, checkIfMail, parseMessage

parser = BytesParser()
parsed = []
for i, msgpath in enumerate(Path(MAILDIR).glob("**/*")):
    if not checkIfMail(msgpath):
        continue
    try:
        msg = parser.parse(open(msgpath, "rb"), headersonly=True)
        parsed.append(parseMessage(msg))
    except:
        pass

addresses = set()
datetimes = []
addressnum = []
for i,(addresslist, date) in enumerate(sorted(parsed, key=lambda x: x[1])):
    if date <= 0:
        continue
    for addr in addresslist:
        addresses.add(addr[1])
    datetimes.append(datetime.fromtimestamp(date))
    addressnum.append(len(addresses))

fig = plt.figure()
plt.plot(datetimes, addressnum, label = "# addresses")
plt.plot(datetimes, [i+1 for i in range(len(datetimes))], label = "# emails")
plt.legend()
fig.autofmt_xdate()
plt.savefig("date-address.svg")


fig = plt.figure()
plt.plot([i+1 for i in range(len(addressnum))], addressnum)
plt.xlabel("# emails")
plt.ylabel("# addresses")
plt.savefig("email-address.svg")


