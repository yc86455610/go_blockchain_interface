package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/levigross/grequests"
)

const min = 0.000000000001

func isEqual(f1, f2 float64) bool {
	if f1 < f2 {
		return isEqual(f2, f1)
	}
	return math.Dim(f1, f2) < min
}

func HTTPGet(url string, requestOptions *grequests.RequestOptions) (response []byte, err error) {
	httpResponse, err := grequests.Get(url, requestOptions)
	if err == nil {
		if httpResponse.StatusCode == 200 {
			response = httpResponse.Bytes()
		}
	}
	return
}

// TronBlocksChainCheck 根据提币的数量，提币方地址以及目标方地址来检查提币是否已经confirmed.
// 返回值有两个：提币状态以及已收到的提币数量（扣除手续费）
func TronBlocksChainCheck(withdrawAmount float64, originalAddress string, targetAddress string) (status string, netWithdrawAmount float64, confirmations int64, err error) {
	url := fmt.Sprintf("http://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999", targetAddress)
	bData, err := HTTPGet(url, nil)
	if err != nil {
		return
	}
	transactions, _, _, err := jsonparser.Get(bData, "result")
	_, err = jsonparser.ArrayEach(transactions, func(value []byte, dataType jsonparser.ValueType, offset int, e error) {
		_from, _, _, e := jsonparser.Get(value, "from")
		_to, _, _, e := jsonparser.Get(value, "to")
		_value, _, _, e := jsonparser.Get(value, "value")
		_txreceiptStatus, _, _, e := jsonparser.Get(value, "txreceipt_status")
		_gasPrice, _, _, e := jsonparser.Get(value, "gasPrice")
		_gasUsed, _, _, e := jsonparser.Get(value, "gasUsed")
		_confirmations, _, _, e := jsonparser.Get(value, "confirmations")

		sf := strings.ToLower(string(_from))
		st := strings.ToLower(string(_to))
		sv := strings.ToLower(string(_value))
		iv, _ := strconv.ParseFloat(sv, 64)
		sts := string(_txreceiptStatus)
		sgp := string(_gasPrice)
		igp, _ := strconv.ParseFloat(sgp, 64)
		sgu := string(_gasUsed)
		igu, _ := strconv.ParseFloat(sgu, 64)
		sc := string(_confirmations)

		if sf == strings.ToLower(originalAddress) && st == strings.ToLower(targetAddress) && isEqual(iv, withdrawAmount) {
			// fmt.Println(sf, st, sv, iv, sgp, igp, sgu, igu, sc)

			// 已完成的提币数量，未扣除提币的手续费
			fmt.Println("gas:", igp*igu)
			// 已收到币的实际数量，扣除了提币的手续费
			fmt.Println("net_receive_amount:", iv)
			if sts == "" || sts == "1" {
				status = "confirmed"
			} else {
				status = "online"
			}
			netWithdrawAmount = iv
			confirmations, _ = strconv.ParseInt(sc, 10, 64)
		} else {
			e = errors.New("this is a new error")
		}
	})
	return
}

func main() {
	status, netReceiveAmount, confirmations, err := TronBlocksChainCheck(1853900000000000, "0xce2f4e757bb0175fef1463dd28da033d29fe2a68", "0x4c61414f6c7955fedc81b1da2d6b149ec762bb3a")
	if err != nil {
		fmt.Println("request failed...")
		return
	}
	fmt.Println(fmt.Sprintf("status: %s, net_withdraw_amount: %f, confirmations: %d", status, netReceiveAmount, confirmations))
}
