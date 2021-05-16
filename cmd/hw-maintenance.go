package main

import (
    "encoding/json"
    "fmt"
    "flag"
    "os"
    "log"
    "net/http"
    "time"
)

const (
    VerdorLenovo        = "Lenovo"
)

var (
    // flag
    verdor          string
    serialNumber    string

    // app info
    AppName             string
    AppAuthor           string
    AppVersion          string
    AppGitCommitHash    string
    AppBuildTime        string
    AppGoVersion        string
)

func init() {
    flag.StringVar(&verdor, "V", "", "verdor name, like Lenovo, Huawei, IBM ...")
    flag.StringVar(&serialNumber, "sn", "", "the serial number of target machine")
    flag.Parse()
}

func main() {
    var err error
    var res string

    if err = checkArgs(); err != nil {
        log.Fatal(err)
    }

    switch verdor {
    case VerdorLenovo:
        res, err = queryLenovo(serialNumber)
    default:
        err = fmt.Errorf("not support the verdor %s now", verdor)
    }

    if err != nil {
        log.Fatal(err)
    }

    fmt.Print(res)
}

func flagUsage() {
    usageMsg := fmt.Sprintf(`App: %s
Version: %s
Author: %s
GitCommit: %s
BuildTime: %s
GoVersion: %s
Options:
`, AppName, AppVersion, AppAuthor, AppGitCommitHash, AppBuildTime, AppGoVersion)

    fmt.Fprintf(os.Stderr, usageMsg)
    flag.PrintDefaults()
}

func checkArgs() (error) {
    if verdor == "" || serialNumber == "" {
        return fmt.Errorf("the input verdor and serialNumber must be valid")
    }

    return nil
}

type lenovoWD struct {
    OnsiteStartDate string              `json:"OnsiteStartDate"`
    OnsiteEndDate   string              `json:"OnsiteEndDate"`
}

type lenovoResp struct {
    Status          int                 `json:"status"`
    WarrantyData    []lenovoWD          `json:"WarrantyData"`
}

func queryLenovo(sn string) (string, error) {
    url1 := "https://think.lenovo.com.cn/service/warranty/repairDeploy.html"
    url2 := "https://think.lenovo.com.cn/service/handlers/WarrantyConfigInfo.ashx"

    client := &http.Client{
        Timeout: 10*time.Second,
    }

    // get cookies from the first request
    resp, err := client.Get(url1)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    req, err := http.NewRequest("GET", url2, nil)
    if err != nil {
        return "", err
    }

    for _, v := range resp.Cookies() {
        req.AddCookie(v)
    }

    q := req.URL.Query()
    q.Add("Method", "WarrantyConfigSearch")
    q.Add("MachineNo", sn)
    q.Add("categoryid", "")
    q.Add("CODEName", "")
    q.Add("SearchNodeCC", "")
    req.URL.RawQuery = q.Encode()

    req.Header.Set("Referer", url1)

    resp, err = client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var respS lenovoResp
    err = json.NewDecoder(resp.Body).Decode(&respS)
    if err != nil {
        return "", err
    }

    data, err := json.Marshal(respS)
    if err != nil {
        return "", err
    }

    return string(data), nil
}