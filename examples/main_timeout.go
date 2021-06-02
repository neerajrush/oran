package main

import (
    "fmt"
    "net"
    "encoding/binary"
    "encoding/json"
    "time"
    "log"
)

const ALARM_MGR = 0

type PanelErr struct {
	PanelNo		int	`json:"panel_no"`
	BandNo		int	`json:"band_no"`
	ChannelNo	int	`json:"channel_no"`
	RficNo		int	`json:"rfic_no"`
}

type AlarmInfo struct{
	SubSystemId	int	`json:"subsystem_id"`
	ErrorCode	int	`json:"error_code"`
	Severity	string	`json:"severity"`
	ErrorCodeStr    string  `json:"error_code_str"`
	ShortDesc       string  `json:"short_desc"`
	BitMap	        int	`json:"bitmap"`
	MainCmd		int	`json:"main_cmd"`
	SubCmd		int	`json:"sub_cmd"`
	OccurredAt	string	`json:"occurred_at"`
	FunctionName	string	`json:"function_name"`
	PanelError	PanelErr`json:"panel_error"`
}

type JReq struct {
	Component 	int  `json:"component"`
	Command   	int  `json:"command"`
	Subcommand   	int  `json:"subcommand"`
}

func NewPanelError(panelNo, bandNo, channelNo, rficNo int) *PanelErr {
	pe := PanelErr {
		PanelNo:	panelNo,
		BandNo:		bandNo,
		ChannelNo:	channelNo,
		RficNo:		rficNo,
		}

	return &pe
}

func NewAlarmInfo(sysId, err int, sev, errStr, shortDsc string, bitmap, mainCmd, subCmd int, occurAt, fnName string,  pe PanelErr) *AlarmInfo {
	alarmInfo := AlarmInfo{
			SubSystemId:	sysId,
			ErrorCode:	err,
			Severity:	sev,
			ErrorCodeStr:	errStr,
			ShortDesc:	shortDsc,
			BitMap:		bitmap,
			MainCmd:	mainCmd,
			SubCmd:		subCmd,
			OccurredAt:	occurAt,
			FunctionName:	fnName,
			PanelError:	pe,
		     }

	     return &alarmInfo
}

func createJsonResponse(cmd string) string {
	resp := "{ \"result\": 0, \"result_str\": \"Success\" }"
	log.Println("Cmd:", cmd)
	if ALARM_MGR == 1 {
		pe := NewPanelError(10, 20, 30, 40)
		alarm := NewAlarmInfo(1, 2, "Major", "TEST_ERROR", "Sample Error", 0x1F, 6, 7, time.Now().String(), "PanelPoll", *pe)
		response,_ := json.Marshal(alarm)
		return string(response)
	}
	return resp;
}

func createBcmConfigResponse() string {
	resp := "{\"panel_type\": \"HALF-PANEL\", \"num_bands\": 2, \"bands\": [{\"band\": \"AWS\", \"config\": \"vita\", \"num_carriers\": 1, \"carriers\": [{\"carrier_num\": 1, \"tx_rf_frequency\": 1842.5, \"rx_rf_frequency\": 1747.5, \"bandwidth\": \"10MHZ\"}]}, {\"band\": \"PCS\", \"config\": \"vita\", \"num_carriers\": 1, \"carriers\": [{\"carrier_num\": 1, \"tx_rf_frequency\": 1842.5, \"rx_rf_frequency\": 1747.5, \"bandwidth\": \"10MHZ\"}]}]}"
	log.Println("Rsp:", resp)
	return resp;
}

func handleClient(conn net.Conn) {
    defer conn.Close()

    var cmdLen int32

    if err := binary.Read(conn, binary.LittleEndian, &cmdLen); err != nil {
        fmt.Println("Error reading:", err.Error())
        return
    }

    fmt.Println("CmdLen:", cmdLen)

    buf := make([]byte, cmdLen)
    _, err := conn.Read(buf)
    if err != nil {
       fmt.Println("Error reading:", err.Error())
       return
    }
    fmt.Println("Received Cmd:", string(buf))

    resp := createJsonResponse(string(buf))
    var respLen uint32
    respLen = uint32(len(resp))
    fmt.Println("Sending Resp (len):", respLen)
    binary.Write(conn, binary.LittleEndian, respLen)
    _, err = conn.Write([]byte(resp))
    if err != nil {
       fmt.Println("Error writing resp:", err.Error())
       return
    }
}

func handleBMCClient(conn net.Conn) {
    defer conn.Close()

    var cmdLen int32

    if err := binary.Read(conn, binary.LittleEndian, &cmdLen); err != nil {
        fmt.Println("Error reading:", err.Error())
        return
    }

    fmt.Println("CmdLen:", cmdLen)

    cmd := make([]byte, cmdLen)
    n := 0
    total_n := int32(0)
    for {
	n, err := conn.Read(cmd[n:])
	if err != nil {
		log.Println("Failed to read:", err)
		return
	}
	total_n += int32(n)
	fmt.Println("Received Request(len):", n, "total:", total_n)
	if total_n >= cmdLen {
		break;
	}
    }

    fmt.Println("Received Cmd:", string(cmd))
    var bcmReq JReq
    err := json.Unmarshal(cmd, &bcmReq)
    if err != nil {
		log.Println("Failed to unmarshal:", err)
		return
    }
    var resp  string
    if bcmReq.Command == 48 && bcmReq.Subcommand == 70 {
    	resp = createBcmConfigResponse()
    } else {
    	resp = createJsonResponse(string(cmd))
    }
    time.Sleep(300 * time.Second)
    fmt.Println("Response:", resp)
    var respLen uint32
    respLen = uint32(len(resp))
    fmt.Println("Sending Resp (len):", respLen)
    binary.Write(conn, binary.LittleEndian, respLen)
    _, err = conn.Write([]byte(resp))
    if err != nil {
       fmt.Println("Error writing resp:", err.Error())
       return
    }
}

func main() {
    listener, err := net.Listen("tcp", ":9999")
    if err != nil {
        fmt.Println("Error listening:", err.Error())
	return
    }
    defer listener.Close()
    fmt.Println("Listening on " + ":9999")
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
	    return
        }
        go handleBMCClient(conn)
    }
}

