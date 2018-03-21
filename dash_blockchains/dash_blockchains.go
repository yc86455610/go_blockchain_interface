package main

import (
	"encoding/json"
	"fmt"
	"math"

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

// DashBlocksChainCheck 根据提币的数量，提币方地址以及目标方地址来检查提币是否已经confirmed.
// 返回值有两个：提币状态以及已收到的提币数量（扣除手续费）
func DashBlocksChainCheck(withrawAmount float64, originalAddress string, targetAddress string) (status string, netWithdrawAmount float64, confirmations int64, err error) {
	targetURL := fmt.Sprintf("https://api.blockcypher.com/v1/dash/main/addrs/%s/full", targetAddress)
	bData, err := HTTPGet(targetURL, nil)
	if err != nil {
		fmt.Println("error: HTTPGet targetURL failed.")
		return
	}
	transactions, _, _, err := jsonparser.Get(bData, "txs")
	if err != nil {
		fmt.Println("error: transaction.")
		return
	}
	// fmt.Println(string(transactions))
	_, err = jsonparser.ArrayEach(transactions, func(value []byte, dataType jsonparser.ValueType, offset int, e error) {
		_outs, _, _, e := jsonparser.Get(value, "outputs")
		_fees, _, _, e := jsonparser.Get(value, "fees")
		_confirmations, _, _, e := jsonparser.Get(value, "confirmations")
		fees, e := jsonparser.GetInt(_fees)
		confirmations, e = jsonparser.GetInt(_confirmations)

		status = "online"
		jsonparser.ArrayEach(_outs, func(out []byte, dataType jsonparser.ValueType, offset int, e error) {
			_addr, _, _, e := jsonparser.Get(out, "addresses")
			_value, _, _, e := jsonparser.Get(out, "value")
			var outsAddr []string
			e = json.Unmarshal(_addr, &outsAddr)
			v, e := jsonparser.GetFloat(_value)
			if outsAddr[0] == targetAddress && v == withrawAmount {
				fmt.Println("fees: ", fees)
				status = "comfirmed"
				netWithdrawAmount = v
			}
		})
	})
	return
}

func main() {
	status, netReceiveAmount, confirmations, err := DashBlocksChainCheck(10000000, "XrBtTsARPgGaqU3RsLtryFxDmaGkNPZdF8", "XnwjLMM2yGeK9tCSAwdkADjiGUaEotXaNV")
	if err != nil {
		fmt.Println("request failed...")
		return
	}
	fmt.Println(fmt.Sprintf("status: %s, net_withdraw_amount: %f, confirmations: %d", status, netReceiveAmount, confirmations))
}
