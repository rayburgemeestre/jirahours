from datetime import date, timedelta

sdate = date(2020, 1,  1)   # start date
edate = date(2020, 3,  1)   # end date

delta = edate - sdate       # as timedelta

for i in range(delta.days + 1):
    day = sdate + timedelta(days=i)
    if day.weekday() >= 5: continue
    print(day)

