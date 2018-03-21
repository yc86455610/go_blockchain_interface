package main

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/levigross/grequests"
)

// HTTPGet .
func HTTPGet(url string, requestOptions *grequests.RequestOptions) (response []byte, err error) {
	httpResponse, err := grequests.Get(url, requestOptions)
	if err == nil {
		if httpResponse.StatusCode == 200 {
			response = httpResponse.Bytes()
		}
	}
	return
}

// RemoveTailZeroCharacter .
func RemoveTailZeroCharacter(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != '0' {
			return s[:i+1]
		}
	}
	return "0"
}

// BtcBlocksChainCheck 根据提币的数量，提币方地址以及目标方地址来检查提币是否已经confirmed.
// 返回值有两个：提币状态以及已收到的提币数量（扣除手续费）
func BtcBlocksChainCheck(withdrawAmount float64, originalAddress string, targetAddress string) (status string, netWithdrawAmount float64, err error) {
	url := fmt.Sprintf("https://blockchain.info/rawaddr/%s", targetAddress)
	bData, err := HTTPGet(url, nil)
	if err != nil {
		return
	}
	transactions, _, _, err := jsonparser.Get(bData, "txs")
	jsonparser.ArrayEach(transactions, func(value []byte, dataType jsonparser.ValueType, offset int, e error) {
		inputs, _, _, e := jsonparser.Get(value, "inputs")
		outs, _, _, e := jsonparser.Get(value, "out")
		var totalIn, totalOut, missedTotalIn float64
		jsonparser.ArrayEach(inputs, func(ipt []byte, dataType jsonparser.ValueType, offset int, e error) {
			prevOut, _, _, e := jsonparser.Get(ipt, "prev_out")
			_addr, _, _, e := jsonparser.Get(prevOut, "addr")
			_spent, _, _, e := jsonparser.Get(prevOut, "spent")
			_value, _, _, e := jsonparser.Get(prevOut, "value")

			a, e := jsonparser.GetBoolean(_spent)
			b := string(_addr)
			c, e := jsonparser.GetFloat(_value)
			if a && b == originalAddress {
				totalIn += c
			}

		})

		jsonparser.ArrayEach(outs, func(out []byte, dataType jsonparser.ValueType, offset int, e error) {
			_addr, _, _, e := jsonparser.Get(out, "addr")
			_spent, _, _, e := jsonparser.Get(out, "spent")
			_value, _, _, e := jsonparser.Get(out, "value")

			a, e := jsonparser.GetBoolean(_spent)
			b := string(_addr)
			c, e := jsonparser.GetFloat(_value)
			if a && b == targetAddress {
				totalOut += c
			}
			if !a && b == originalAddress {
				missedTotalIn += c
			}
		})

		if totalIn > 0 && totalOut > 0 {
			netTotalIn := totalIn - missedTotalIn

			strWithdrawAmount := strconv.FormatFloat(withdrawAmount*1e15, 'f', 0, 64)
			strNetTotalIn := strconv.FormatFloat(netTotalIn*1e5, 'f', 0, 64)
			strTotalOut := strconv.FormatFloat(totalOut*1e5, 'f', 0, 64)

			strWithdrawAmount = RemoveTailZeroCharacter(strWithdrawAmount) // 去除字符串尾部的0字符
			strNetTotalIn = RemoveTailZeroCharacter(strNetTotalIn)
			strTotalOut = RemoveTailZeroCharacter(strTotalOut)

			floatWithdrawAmount, _ := strconv.ParseFloat(strWithdrawAmount, 64)
			floatNetTotalIn, _ := strconv.ParseFloat(strNetTotalIn, 64)
			floatTotalOut, _ := strconv.ParseFloat(strTotalOut, 64)
			scale := floatWithdrawAmount / withdrawAmount
			finishedWithdrawAmount := floatNetTotalIn / scale
			netReceiveAmount := floatTotalOut / scale

			// 已完成的提币数量，未扣除提币的手续费
			fmt.Println("finished_withdraw_amount:", finishedWithdrawAmount)
			// 已收到币的实际数量，扣除了提币的手续费
			fmt.Println("net_receive_amount:", netReceiveAmount)
			if withdrawAmount == finishedWithdrawAmount {
				status = "confirmed"
			} else {
				status = "online"
			}
			netWithdrawAmount = netReceiveAmount
			return
		}
	})
	return
}

func main() {
	status, netReceiveAmount, err := BtcBlocksChainCheck(0.04907017, "3BMEXebzKg6WbeUiGmZ8n8rzhu4Axutaf2", "35TWgWwVeU7mu6JZvGzRgwx1W6dw4F6xXx")
	if err != nil {
		fmt.Println("request failed...")
	} else {
		fmt.Println(fmt.Sprintf("status: %s, net_withdraw_amount: %f", status, netReceiveAmount))
	}
}
