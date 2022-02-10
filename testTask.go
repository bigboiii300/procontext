package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Valute  []struct {
		Nominal string `xml:"Nominal"`
		Name    string `xml:"Name"`
		Value   string `xml:"Value"`
	} `xml:"Valute"`
}

var valueMax float64 = 0
var nameMax string
var dateMax string
var nominalMax int

var valueMin float64 = math.MaxInt32
var nameMin string
var dateMin string
var nominalMin int

func main() {
	currentDate := time.Now()
	fmt.Printf("Текущая дата: %v\n", currentDate)
	showLogs := chooseCommand()
	period := 90
	tempDate := currentDate
	avgValute := fillStruct(currentDate.Format("02/01/2006"))
	for i := 0; i < period; i++ {
		getBackDay := tempDate.AddDate(0, 0, -1)
		currentTime := getBackDay.Format("02/01/2006")
		tempDate = getBackDay
		v := parser(currentTime)
		if showLogs {
			fmt.Printf("%q\n%q\n\n", v.Date, v.Valute)
		}
		for i := 0; i < len(v.Valute); i++ {
			valueStr := strings.Replace(v.Valute[i].Value, ",", ".", -1)
			nominalStr := strings.Replace(v.Valute[i].Nominal, ",", ".", -1)
			if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
				if nominalTemp, err := strconv.Atoi(nominalStr); err == nil {
					valueOneCurrency := val / float64(nominalTemp)
					valueMax, nameMax, dateMax, nominalMax =
						getMax(valueMax, valueOneCurrency, val, nameMax, v.Valute[i], dateMax, v, nominalMax, nominalTemp)
					valueMin, nameMin, dateMin, nominalMin =
						getMin(valueMin, valueOneCurrency, nameMin, v.Valute[i], dateMin, v, nominalMin, nominalTemp)
					valAvg, _ := strconv.ParseFloat(avgValute.Valute[i].Value, 64)
					avgValute.Valute[i].Value = fmt.Sprintf("%.4f", valAvg+val)
				}
			}
		}
	}
	getAvg(avgValute, period)
	output(nameMax, valueMax, dateMax, nominalMax, nameMin, valueMin, nominalMin, dateMin, avgValute)
}

func chooseCommand() bool {
	fmt.Printf("Введите одну из двух команд:\n")
	fmt.Printf("1 - Вывод всех логов и результата.\n")
	fmt.Printf("2 - Вывод результата без логов.\n")
	fmt.Printf("При некорректном вводе автоматически выбирается вторая команда.\n")
	var inputForLogs int
	var showLogs bool
	fmt.Scanf("%d", &inputForLogs)
	switch inputForLogs {
	case 1:
		showLogs = true
	case 2:
		showLogs = false
	default:
		fmt.Printf("Неверно введена команда! Автоматически выбран режим без показа логов\n")
	}
	return showLogs
}

func getAvg(avgValute ValCurs, period int) {
	for i := 0; i < len(avgValute.Valute); i++ {
		valueF, _ := strconv.ParseFloat(avgValute.Valute[i].Value, 64)
		avg := valueF / float64(period)
		avgValute.Valute[i].Value = fmt.Sprintf("%.4f", avg)
	}
}

func fillStruct(currentTime string) ValCurs {
	var avgValute ValCurs
	avgValute = parser(currentTime)
	for _, valute := range avgValute.Valute {
		valute.Value = "0"
	}
	return avgValute
}

func output(nameMax string, valueMax float64, dateMax string, nominalMax int, nameMin string,
	valueMin float64, nominalMin int, dateMin string, tempValute ValCurs) {
	fmt.Printf("Максимальный курс: \n")
	fmt.Printf("Название валюты: %s \n", nameMax)
	fmt.Printf("Курс: %f\n", valueMax)
	fmt.Printf("Дата: %s\n", dateMax)
	fmt.Printf("Номинал валюты: %v", nominalMax)

	fmt.Printf("\n\nМинимальный курс:\n")
	fmt.Printf("Название валюты: %s\n", nameMin)
	fmt.Printf("Курс: %f\n", valueMin*float64(nominalMin))
	fmt.Printf("Дата: %s\n", dateMin)
	fmt.Printf("Номинал валюты: %v", nominalMin)

	fmt.Printf("\n\nСреднее значение курса рубля по всем валютам:\n\n")
	for _, valute := range tempValute.Valute {
		fmt.Printf("Название валюты: %s\n", valute.Name)
		fmt.Printf("Средний курс: %s\n", valute.Value)
		fmt.Printf("Номинал валюты: %s\n\n", valute.Nominal)
	}
}

func getMin(valueMin float64, valueOneCurrency float64, nameMin string,
	value struct {
		Nominal string `xml:"Nominal"`
		Name    string `xml:"Name"`
		Value   string `xml:"Value"`
	}, dateMin string, v ValCurs, nominalMin int, nominalTemp int) (float64, string, string, int) {
	if valueMin > valueOneCurrency {
		valueMin = valueOneCurrency
		nameMin = value.Name
		dateMin = v.Date
		nominalMin = nominalTemp
	}
	return valueMin, nameMin, dateMin, nominalMin
}

func getMax(valueMax float64, valueOneCurrency float64, val float64, nameMax string, value struct {
	Nominal string `xml:"Nominal"`
	Name    string `xml:"Name"`
	Value   string `xml:"Value"`
}, dateMax string, v ValCurs, nominalMax int, nominalTemp int) (float64, string, string, int) {
	if valueMax < valueOneCurrency {
		valueMax = val
		nameMax = value.Name
		dateMax = v.Date
		nominalMax = nominalTemp
	}
	return valueMax, nameMax, dateMax, nominalMax
}

func parser(currentTime string) ValCurs {
	resp, _ := http.Get("https://www.cbr.ru/scripts/XML_daily_eng.asp?date_req=" + currentTime)
	bytes, _ := ioutil.ReadAll(resp.Body)
	test := strings.Replace(string(bytes), "<?xml version=\"1.0\" encoding=\"windows-1251\"?>", "", -1)
	_ = resp.Body.Close()
	var v ValCurs
	_ = xml.Unmarshal([]byte(test), &v)
	return v
}
