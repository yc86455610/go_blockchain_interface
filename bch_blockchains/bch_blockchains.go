package main

import (
	"encoding/json"
	"fmt"

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

// BchBlocksChainCheck 根据提币的数量，提币方地址以及目标方地址来检查提币是否已经confirmed.
// 返回值有两个：提币状态以及已收到的提币数量（扣除手续费）
func BchBlocksChainCheck(withdrawAmount float64, originalAddress string, targetAddress string) (status string, netWithdrawAmount float64, confirmations int64, err error) {
	targetURL := fmt.Sprintf("https://bch-chain.api.btc.com/v3/address/%s/tx", targetAddress)
	bData, err := HTTPGet(targetURL, nil)
	if err != nil {
		fmt.Println("error: HTTPGet targetURL failed.")
		return
	}
	//fmt.Println(string(bData))

	_, err = jsonparser.ArrayEach(bData, func(value []byte, dataType jsonparser.ValueType, offset int, e error) {
		_outs, _, _, e := jsonparser.Get(value, "outputs")
		_confirmations, _, _, e := jsonparser.Get(value, "confirmations")
		_fees, _, _, e := jsonparser.Get(value, "fee")
		fees, e := jsonparser.GetInt(_fees)
		confirmations, e = jsonparser.GetInt(_confirmations)

		status = "online"
		jsonparser.ArrayEach(_outs, func(out []byte, dataType jsonparser.ValueType, offset int, e error) {
			_addr, _, _, e := jsonparser.Get(out, "addresses")
			_value, _, _, e := jsonparser.Get(out, "value")

			var outputs []string
			e = json.Unmarshal(_addr, &outputs)
			v, e := jsonparser.GetFloat(_value)

			if outputs[0] == targetAddress && v == withdrawAmount {
				fmt.Println("fees: ", fees)
				status = "confirmed"
				netWithdrawAmount = v
			}
		})
	})
	return
}

func main() {
	status, netReceiveAmount, confirmations, err := BchBlocksChainCheck(1567646685, "1JnCVng5JaVSmPxtm3wjkUXVDv2HXZBJJw", "15um5NFHM39xPdnnxwbEY4LeSMC9CE9cky")
	if err != nil {
		fmt.Println("request failed...")
		return
	}
	fmt.Println(fmt.Sprintf("status: %s, net_withdraw_amount: %f, confirmations: %d", status, netReceiveAmount, confirmations))
}

