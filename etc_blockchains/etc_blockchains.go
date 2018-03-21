package main

import (
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

// EtcBlocksChainCheck 根据提币的数量，提币方地址以及目标方地址来检查提币是否已经confirmed.
// 返回值有两个：提币状态以及已收到的提币数量（扣除手续费）
func EtcBlocksChainCheck(withrawAmount float64, originalAddress string, targetAddress string) (status string, netWithdrawAmount float64, confirmations int64, err error) {
	status = "online"
	targetURL := fmt.Sprintf("https://etcchain.com/api/v1/getTransactionsByAddress?address=%s&sort=asc", targetAddress)
	bData, err := HTTPGet(targetURL, nil)
	if err != nil {
		fmt.Println("error: HTTPGet targetURL failed.")
		return
	}
	// fmt.Println(string(bData))

	_, err = jsonparser.ArrayEach(bData, func(value []byte, dataType jsonparser.ValueType, offset int, e error) {
		_from, _, _, e := jsonparser.Get(value, "from")
		_to, _, _, e := jsonparser.Get(value, "to")
		_value, _, _, e := jsonparser.Get(value, "valueEther")
		_confirmations, _, _, e := jsonparser.Get(value, "confirmations")
		_fees, _, _, e := jsonparser.Get(value, "feeEther")

		from := string(_from)
		to := string(_to)
		v, e := jsonparser.GetFloat(_value)
		fees, e := jsonparser.GetFloat(_fees)

		if from == originalAddress && to == targetAddress && v == withrawAmount {
			fmt.Println("fees: ", fees)
			status = "comfirmed"
			netWithdrawAmount = v
			confirmations, e = jsonparser.GetInt(_confirmations)
			//fmt.Println(status)
			return
		}
	})
	return
}

func main() {
	status, netReceiveAmount, confirmations, err := EtcBlocksChainCheck(0.931847144, "0x6b83f808fce08f51adb2e9e35a21a601e702785f", "0x4f0d74ecdceb209213ecc0cce6366c3bef1fce57")
	if err != nil {
		fmt.Println("request failed...")
		return
	}
	fmt.Println(fmt.Sprintf("status: %s, net_withdraw_amount: %f, confirmations: %d", status, netReceiveAmount, confirmations))
}
