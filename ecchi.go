package main

import (
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "sync"
)

var (
    root        = ""
    miumBaseUrl = "http://blog.livedoor.jp/kirekawa39-siro/archives/300mium-%03d.html"
    siroBaseUrl = "http://blog.livedoor.jp/kirekawa39-siro/archives/siro-%04d.html"
    ganaBaseUrl = "http://blog.livedoor.jp/kirekawa39-siro/archives/200GANA-%d.html"
    luxuBaseUrl = "http://blog.livedoor.jp/kirekawa39-siro/archives/259LUXU-%3d.html"
    dcvBaseUrl  = "http://blog.livedoor.jp/kirekawa39-siro/archives/277dcv-%03d.html"
    regex       *regexp.Regexp
    wg          sync.WaitGroup
)

func init() {
    cdir, _ := os.Getwd()
    root = filepath.Join(cdir, "images")
    regex, _ = regexp.Compile(`[0-9a-zA-z]+-\d+`)
}

func main() {
    args := os.Args
    url := ""
    if len(args[1:]) == 3 {
        switch args[1] {
        case "-m":
            url = miumBaseUrl
        case "-s":
            url = siroBaseUrl
        case "-g":
            url = ganaBaseUrl
        case "-l":
            url = luxuBaseUrl
        case "-d":
            url = dcvBaseUrl
        default:
            errorMsg()
            return
        }
        start, _ := strconv.Atoi(args[2])
        end, _ := strconv.Atoi(args[3])
        for i := start; i <= end; i++ {
            wg.Add(1)
            go fetch(url, i)
        }
    } else {
        errorMsg()
    }
    wg.Wait()
}

func errorMsg() {
    message := "Usage: \n" +
        "\tSIRO => ./ecchi -s  START_NUM END_NUM\n" +
        "\tMIUM => ./ecchi -m  START_NUM END_NUM\n" +
        "\tGANA => ./ecchi -g  START_NUM END_NUM\n" +
        "\tluxu => ./ecchi -l  START_NUM END_NUM\n" +
        "\tdcv  => ./ecchi -d  START_NUM END_NUM\n"
    fmt.Println(message)
}

func fetch(baseUrl string, number int) {
    url := fmt.Sprintf(baseUrl, number)
    dirName := regex.FindString(url)
    dirPath := filepath.Join(root, dirName)

    res, err := http.Get(url)
    if err != nil {
        log.Fatal(err)
    }
    if res.StatusCode == 200 {
        os.MkdirAll(dirPath, 0755)
        doc, err := goquery.NewDocumentFromResponse(res)
        if err != nil {
            log.Fatal(err)
        }
        doc.Find("a").Each(func(i int, s *goquery.Selection) {
            if imageUrl, ok := s.Attr("href"); ok {
                if strings.HasSuffix(imageUrl, ".jpg") {
                    fileName := getFileName(imageUrl)
                    fmt.Printf("[%s] Downloading : %s...\n", dirName, fileName)
                    filePath := filepath.Join(dirPath, fileName)
                    resp, _ := http.Get(imageUrl)
                    fp, _ := os.Create(filePath)
                    io.Copy(fp, resp.Body)
                    resp.Body.Close()
                    fp.Close()
                }
            }
        })
        bq := doc.Find("blockquote")
        fp, _ := os.Create(filepath.Join(dirPath, "description.txt"))
        fp.WriteString(bq.Text())
        fp.Close()
    } else {
        fmt.Printf("[%s] Not found from this site\n", dirName)
    }
    wg.Done()
}

func getFileName(path string) string {
    stringArr := strings.Split(path, "/")
    fileName := stringArr[len(stringArr)-1]
    return fileName
}
