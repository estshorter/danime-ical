# danime-ial

[dアニメストア](https://anime.dmkt-sp.jp)の[今季アニメ一覧ページ](https://anime.dmkt-sp.jp/animestore/CF/winter)の情報をもとに、指定したアニメの放送情報を表すiCalデータを作るプログラムです。
アニメ一覧ページのURLおよび、アニメタイトルの指定は`configs.json`で行います。
`config.json`の場所は下記のようにコマンドライン引数で与えます。
指定しない場合には、カレントディレクトリのものが読み込まれます。

`danime-ical.exe PATH_TO_CONFIGS_JSON`

生成されるiCalファイルの例はこちら。

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