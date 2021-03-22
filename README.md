# danime-ical

[dアニメストア](https://anime.dmkt-sp.jp)の[今季アニメ一覧ページ](https://anime.dmkt-sp.jp/animestore/CF/spring)の情報をスクレイピングし、指定したアニメの配信情報を表すiCalデータを作るプログラム。iCalデータは手動で[Googleカレンダーに読み込ませる](https://support.google.com/calendar/answer/37118)ことを想定している。アニメ一覧ページのURLおよび、アニメタイトルの指定は`configs.json`で行う。`configs.json`の場所は下記のようにコマンドライン引数で与える。指定しない場合には、カレントディレクトリのものが読み込まれる。

`danime-ical.exe PATH_TO_CONFIGS_JSON`

生成されるiCalファイルの例は以下の通り。クールの初めの月（冬アニメだったら1月など）の最初の配信曜日から、毎週の配信予定が記述されている。アニメの話数は13話を想定。

``` ics
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Arran Ubels//Golang ICS library 
METHOD:REQUEST
BEGIN:VEVENT
UID:のんのんびより のんすとっぷ
DTSTART;TZID=Asia/Tokyo:20210101T010000
DTEND;TZID=Asia/Tokyo:20210101T013000
SUMMARY:のんのんびより のんすとっぷ
RRULE:FREQ=WEEKLY;COUNT=13
END:VEVENT
END:VCALENDAR
```
## Requirement
- Chrome
- [ChromeDriver](https://chromedriver.chromium.org/)
