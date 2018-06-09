package core

import (
	asura "github.com/teragrid/asura/types"
	ctypes "github.com/teragrid/teragrid/rpc/core/types"
	"github.com/teragrid/teragrid/version"
	cmn "github.com/teragrid/teralibs/common"
)

// Query the application for some information.
//
// ```shell
// curl 'localhost:46657/asura_query?path=""&data="abcd"&prove=true'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:46657", "/websocket")
// result, err := client.asuraQuery("", "abcd", true)
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
// {
// 	"error": "",
// 	"result": {
// 		"response": {
// 			"log": "exists",
// 			"height": 0,
// 			"proof": "010114FED0DAD959F36091AD761C922ABA3CBF1D8349990101020103011406AA2262E2F448242DF2C2607C3CDC705313EE3B0001149D16177BC71E445476174622EA559715C293740C",
// 			"value": "61626364",
// 			"key": "61626364",
// 			"index": -1,
// 			"code": 0
// 		}
// 	},
// 	"id": "",
// 	"jsonrpc": "2.0"
// }
// ```
//
// ### Query Parameters
//
// | Parameter | Type   | Default | Required | Description                                    |
// |-----------+--------+---------+----------+------------------------------------------------|
// | path      | string | false   | false    | Path to the data ("/a/b/c")                    |
// | data      | []byte | false   | true     | Data                                           |
// | height    | int64 | 0       | false    | Height (0 means latest)                        |
// | trusted   | bool   | false   | false    | Does not include a proof of the data inclusion |
func asuraQuery(path string, data cmn.HexBytes, height int64, trusted bool) (*ctypes.ResultasuraQuery, error) {
	resQuery, err := proxyAppQuery.QuerySync(asura.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  !trusted,
	})
	if err != nil {
		return nil, err
	}
	logger.Info("asuraQuery", "path", path, "data", data, "result", resQuery)
	return &ctypes.ResultasuraQuery{*resQuery}, nil
}

// Get some info about the application.
//
// ```shell
// curl 'localhost:46657/asura_info'
// ```
//
// ```go
// client := client.NewHTTP("tcp://0.0.0.0:46657", "/websocket")
// info, err := client.asuraInfo()
// ```
//
// > The above command returns JSON structured like this:
//
// ```json
// {
// 	"error": "",
// 	"result": {
// 		"response": {
// 			"data": "{\"size\":3}"
// 		}
// 	},
// 	"id": "",
// 	"jsonrpc": "2.0"
// }
// ```
func asuraInfo() (*ctypes.ResultasuraInfo, error) {
	resInfo, err := proxyAppQuery.InfoSync(asura.RequestInfo{version.Version})
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultasuraInfo{*resInfo}, nil
}
