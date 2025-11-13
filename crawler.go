package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

/*
Yayın Tipleri:
T: Televizyon
R: Radyo

T1 - Ulusal TV
T2 - Bölgesel TV
T3 - Yerel TV
R1 - Ulusal Radyo
R2 - Bölgesel Radyo
R3 - Yerel Radyo

Kaynaklar:
İnternet Yayın Lisansı Olan Kuruluşlar Listesi (RD/TV/İBYH) - https://www.rtuk.gov.tr/internet-yayin-lisansi-olan-kuruluslar-listesi-rd-tv-ibyh/1758 - https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_internet.php
İnternet Ortamından Yayın İzni Olan Kuruluşlar (RD/TV/İBYH) - https://www.rtuk.gov.tr/internet-ortamindan-yayin-izni-olan-kuruluslar-rd-tv-ibyh/1759 - https://yayinci.rtuk.gov.tr/izintahsisv2/internet_lisans_sorgulama.php
Karasal Ortamdan Lisans Başvurusu Olan Kuruluşlar Listesi (RD/TV) - https://www.rtuk.gov.tr/karasal-ortamdan-lisans-basvurusu-olan-kuruluslar-listesi-rd-tv/1760 - https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_karasal.php
Kablo Yayın Lisansı Olan Kuruluşlar Listesi (RD/TV) - https://www.rtuk.gov.tr/kablo-yayin-lisansi-olan-kuruluslar-listesi-rd-tv/1761 - https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_kablo.php
Uydu Yayın Lisansı Olan Kuruluşlar Listesi (RD/TV) - https://www.rtuk.gov.tr/uydu-yayin-lisansi-olan-kuruluslar-listesi-rd-tv/1762 - https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_uydu.php

İnternet Ortamında Yayın İletim Yetkisi Bulunan Kuruluşlar - https://www.rtuk.gov.tr/internet-ortaminda-yayin-iletim-yetkisi-bulunan-kuruluslar/4007 - https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_platform.php
*/

type InternetStreamLicense struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	FullName    string `json:"fullname"`
	License     string `json:"license"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	LicenseType string `json:"license_type"`
}

type SatelliteBroadcastingLicense struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	License           string `json:"license"`
	LicenseType       string `json:"license_type"`
	LisenceDetail     string `json:"lisence_detail"`
	BroadcastBranding string `json:"broadcast_branding"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	Address           string `json:"address"`
}

// Static and dynamic Variables
var (
	InternetStreamLicenseAPI                  = "https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_sonuc_platform.php"
	InternetStreamLicensefileName             = "./datas/internetStreamLicense.json"
	SatelliteBroadcastingLicenseAPI           = "https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_sonuc.php"
	SatelliteBroadcastingLicenseTVfileName    = "./datas/satelliteBroadcastingLicenseTV.json"
	SatelliteBroadcastingLicenseRadiofileName = "./datas/satelliteBroadcastingLicenseRadio.json"
)

func ReplaceAll(s, old, new string) string {
	// Replace All for Regexp
	if strings.Contains(old, ")") {
		old = strings.ReplaceAll(old, ")", "\\)")
	}
	if strings.Contains(old, "(") {
		old = strings.ReplaceAll(old, "(", "\\(")
	}
	re := regexp.MustCompile(old)
	return re.ReplaceAllString(s, new)
}

func ClearString(s string) string {
	return strings.TrimSpace(ReplaceAll(ReplaceAll(ReplaceAll(ReplaceAll(s, "  ", " "), "\n", ""), "\t", ""), "\r", ""))
}

// SendRequest makes an HTTP request with the specified method, URL, and payload (optional).
func SendRequest(method, url string, payload *bytes.Buffer) (*http.Response, string, error) {
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, "", errors.New("failed to create HTTP request: " + err.Error())
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "tr,en-US;q=0.9,en;q=0.8,tr-TR;q=0.7,zh-CN;q=0.6,zh-TW;q=0.5,zh;q=0.4,ja;q=0.3,ko;q=0.2,bg;q=0.1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://yayinci.rtuk.gov.tr")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

	if payload != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", errors.New("failed to execute request: " + err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.New("failed to read response body: " + err.Error())
	}

	bodyStr := ReplaceAll(ReplaceAll(string(body), "<!--", ""), "-->", "")
	body = []byte(bodyStr)

	return resp, string(body), nil
}

// Get internet stream license from RTUK - https://yayinci.rtuk.gov.tr/izintahsisv2/web_giris_platform.php

func getInternetStreamLicense(fileName string) {
	/*
		<table class="table table-bordered table-condensed table-hover">
		<thead>
		<tr>
			<th></th>
			<th>Ünvanı</th>
			<th>Web URL</th>
			<th>Marka Adı</th>
			<th>Lisans</th>
			<th>Başlama Tarihi</th>
			<th>Bitiş Tarihi</th>
			<th>Lisans Türü</th>
		</tr>
		</thead>
		<tbod>
		<tr>
			<td class="text-center">
				<label>1</label>
			</td>
			<td>
				<span title="AGENA TELEVİZYON YAYINCILIK ANONİM ŞİRKETİ">AGENA TELEVİZYON YAYINCILIK ANONİM ŞİRKE...</span>
			</td>
			<td> https://www.beinconnect.com.tr/ https://www.digiturkplay.com.tr/ </td>
			<td>	</td>
			<td> Platform </td>
			<td>	</td>
			<td>	</td>
			<td>	(İNTERNET)                </td>
		</tr>
		<tr>
			<td class="text-center">
				<label>2</label>
			</td>
			<td>
				<span title="BSN RADYO TELEVİZYON YAYINCILIĞI ANONİM ŞİRKETİ">BSN RADYO TELEVİZYON YAYINCILIĞI ANONİM ...</span>
			</td>
			<td>	www.powerapp.com.tr   </td>
			<td>	</td>
			<td> Platform </td>
			<td>	</td>
			<td>	</td>
			<td>	(İNTERNET)                </td>
		</tr>
		</tbody>
		</table>
	*/

	var licenses []InternetStreamLicense

	payload := bytes.NewBufferString("LisansTipi=Internet")

	_, body, err := SendRequest("POST", InternetStreamLicenseAPI, payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(body)))
	if err != nil {
		log.Fatal(err)
	}

	// Find the table
	bodySelection := doc.Find("table.table.table-bordered.table-condensed.table-hover > tbody")

	// Iterate over the rows
	bodySelection.Find("tr").Each(func(i int, row *goquery.Selection) {
		// Find the columns
		cols := row.Find("td")

		// Create a license object
		license := InternetStreamLicense{}

		// Iterate over the columns
		cols.Each(func(j int, col *goquery.Selection) {

			switch j {
			case 0: // ID
				license.Id, _ = strconv.Atoi(col.Find("label").Text())
			case 1: // Name
				license.Name = ClearString(col.Find("span").AttrOr("title", ""))
			case 2: // URL
				license.Url = ClearString(col.Text())
			case 3: // Full Name
				license.FullName = ClearString(col.Text())
			case 4: // License
				license.License = ClearString(col.Text())
			case 5: // Start Date
				license.StartDate = ClearString(col.Text())
			case 6: // End Date
				license.EndDate = ClearString(col.Text())
			case 7: // License Type
				license.LicenseType = ClearString(col.Text())
				license.LicenseType = ReplaceAll(ReplaceAll(license.LicenseType, "(", ""), ")", "")
			case 8: // License Type
				license.LicenseType = ClearString(col.Text())
				license.LicenseType = ReplaceAll(ReplaceAll(license.LicenseType, "(", ""), ")", "")
			}

			if license.LicenseType == "" {
				license.LicenseType = "İnternet"
			}
		})

		// Append to the list
		licenses = append(licenses, license)
	})

	// Write to the file
	data, err := json.MarshalIndent(licenses, "", "    ")
	if err != nil {
		log.Fatalf("Cannot encode to JSON: %s\n", err)
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to create file: %s\n", err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Failed to close file: %s\n", err)
		}
	}()

	_, err = file.Write(data)
	if err != nil {
		log.Fatalf("Failed to write to file: %s\n", err)
		return
	}

	fmt.Printf("Licenses written to %s\n", fileName)
}

func parseSatelliteBroadcastingLicense(body string) []SatelliteBroadcastingLicense {

	var licenses []SatelliteBroadcastingLicense

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}

	// Find the table
	bodySelection := doc.Find("table.table.table-bordered.table-condensed.table-hover > tbody")

	//dateRegex := regexp.MustCompile(`(?s)<td>(.*?)</td>.*?<td>(.*?)</td>`)

	bodySelection.Find("tr").Each(func(i int, s *goquery.Selection) {
		license := SatelliteBroadcastingLicense{}

		if s.Find("td").Length() > 1 {
			//println(ClearString(s.Find("td").Eq(11).Text()))
			license.Id, _ = strconv.Atoi(s.Find("td").Eq(0).Find("label").Text())
			license.Name = ClearString(s.Find("td").Eq(1).Find("span").AttrOr("title", ""))
			license.License = ClearString(s.Find("td").Eq(2).Find("span").Text())
			license.BroadcastBranding = ClearString(s.Find("td").Eq(3).Text())
			license.LicenseType = ClearString(s.Find("td").Eq(4).Text())
			license.LisenceDetail = ClearString(s.Find("td").Eq(-1).Text())
			license.StartDate = ClearString(s.Find("td").Eq(6).Find("table tr td").Text())
			license.EndDate = ClearString(s.Find("td").Eq(8).Text())
		}

		if s.Find("td small").Length() > 1 {
			license.Address = ClearString(s.Find("td small i").Text())
		}

		licenses = append(licenses, license)
	})

	// Boş olan alanları temizle
	var filteredLicenses []SatelliteBroadcastingLicense
	for _, license := range licenses {
		if license.Name != "" || license.Address != "" || license.Id != 0 {
			filteredLicenses = append(filteredLicenses, license)
		}
	}
	licenses = filteredLicenses

	return licenses
}

// Get Satellite Broadcasting License
func getSatelliteBroadcastingLicense(fileNameTV string, fileNameRadio string) {
	var licensesRadio []SatelliteBroadcastingLicense
	var licensesTV []SatelliteBroadcastingLicense

	payloadRadio := bytes.NewBufferString("YayinTuru=Uydu&VericiTipi=RADYO")
	payloadTV := bytes.NewBufferString("YayinTuru=Uydu&VericiTipi=TV")

	_, bodyRadio, err := SendRequest("POST", SatelliteBroadcastingLicenseAPI, payloadRadio)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, bodyTV, err := SendRequest("POST", SatelliteBroadcastingLicenseAPI, payloadTV)
	if err != nil {
		fmt.Println(err)
		return
	}

	licensesRadio = parseSatelliteBroadcastingLicense(bodyRadio)
	licensesTV = parseSatelliteBroadcastingLicense(bodyTV)

	// Write to the file for Radio
	RadioData, err := json.MarshalIndent(licensesRadio, "", "    ")
	if err != nil {
		log.Fatalf("Cannot encode to JSON: %s\n", err)
		return
	}

	fileRadio, err := os.Create(fileNameRadio)
	if err != nil {
		log.Fatalf("Failed to create file: %s\n", err)
		return
	}

	defer func() {
		if err := fileRadio.Close(); err != nil {
			log.Fatalf("Failed to close file: %s\n", err)
		}
	}()

	_, err = fileRadio.Write(RadioData)
	if err != nil {
		log.Fatalf("Failed to write to file: %s\n", err)
		return
	}

	fmt.Printf("Licenses written to %s\n", fileNameRadio)

	// Write to the file for TV
	TVData, err := json.MarshalIndent(licensesTV, "", "    ")
	if err != nil {
		log.Fatalf("Cannot encode to JSON: %s\n", err)
		return
	}

	fileTV, err := os.Create(fileNameTV)
	if err != nil {
		log.Fatalf("Failed to create file: %s\n", err)
		return
	}

	defer func() {
		if err := fileTV.Close(); err != nil {
			log.Fatalf("Failed to close file: %s\n", err)
		}
	}()

	_, err = fileTV.Write(TVData)
	if err != nil {
		log.Fatalf("Failed to write to file: %s\n", err)
		return
	}

	fmt.Printf("Licenses written to %s\n", fileNameTV)
}

func main() {

	// Check Folder
	if _, err := os.Stat("./datas"); os.IsNotExist(err) {
		os.Mkdir("./datas", os.ModePerm)
	}

	// Get internet stream license from RTUK
	getInternetStreamLicense(InternetStreamLicensefileName)

	// Get Satellite Broadcasting License
	getSatelliteBroadcastingLicense(SatelliteBroadcastingLicenseTVfileName, SatelliteBroadcastingLicenseRadiofileName)

}
