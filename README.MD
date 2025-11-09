# RTÜK Lisanslar Crawler for Golang



It is recommended that people who will use it for data processing use the JSON file. The JSON file is updated every every day at midnight (00:00) for UTC. The data is taken from the site [Anıt Sayaç](https://anitsayac.com/).

## ❗ Important Information

- It is known that errors occur due to incorrect data entry in data with IDs 70, 69, 114, 78, 71, 511, 605, 607, 3580, 1051, 934, 675, 617, 616, 611, 599, 138.
- Blank data appears that way because it is not available on the site.

## Installation

```bash
go mod tidy
go build -o crawler crawler.go
```

## Usage

```bash
go run .
```

## Data Source

- [RTÜK Companies Lists](https://www.rtuk.gov.tr/medya-hizmet-saglayicilar-1757)

## License

[MIT](LISENCE)

## Author

[Ramazan Sancar](https://www.github.com/ramazansancar)
