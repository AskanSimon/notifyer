package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

var re = regexp.MustCompile(`(\d{1,2})\sv\.\s(\d{1,2})`)

func getStarts(dir string) (string, string, string, error) {
	// create context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(dir),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// javascript runtime
	var res *runtime.RemoteObject

	var result string
	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		network.Enable(),
		emulation.SetDeviceMetricsOverride(1700, 0, 1.0, false),
		chromedp.Navigate("https://www.XXXX.de"),

		// click basic cookie
		chromedp.WaitVisible("#cookiehinweisinner1text"),
		chromedp.Click("#cookie_optin", chromedp.NodeVisible),

		// submit cookie banner
		chromedp.WaitVisible("#cookieokouter"),
		chromedp.EvaluateAsDevTools("set_cookie_ok()", &res),
		chromedp.Sleep(2*time.Second),

		// login (tricky submit name)
		chromedp.SetValue("#username", "XXXX", chromedp.ByID),
		chromedp.SetValue("#pw", "XXXX", chromedp.ByID),
		chromedp.EvaluateAsDevTools("document.getElementById('login').submit.click()", &res),
		chromedp.Sleep(2*time.Second),

		// get data
		chromedp.Text("#wk_days", &result, chromedp.NodeReady, chromedp.ByID),
		chromedp.Nodes(`#wk_days > div > div > div`, &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		return "", "", "", err
	}
	startInfo := strings.TrimSpace(result)

	// filter available data
	matches := re.FindAllStringSubmatch(startInfo, -1)
	var b bytes.Buffer
	for _, v := range matches {
		b.WriteString(v[1])
	}
	starts := b.String()

	// filter status
	var bb bytes.Buffer
	for _, n := range nodes {
		class := n.AttributeValue("class")
		if len(class) > 20 && class[:19] == "wk_status" {
			bb.WriteString(class[20:] + "\n")
		}
	}
	status := bb.String()

	return startInfo, starts, status, nil
}

// sendMail is just used for logging
func sendMail(subject, body string) {
	to := "3d89S_notifiyer@gmail.com"
	from := "XXXX@XXXX.de"
	pass := "XXXX"
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)
	err := smtp.SendMail("mail.XXXX.de:587", smtp.PlainAuth("", from, pass, "mail.XXXX.de"), from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("smtp error: %s\n", err)
		return
	}
}

var lastStarts string
var lastStatus string

func main() {

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)
	}
	//defer os.RemoveAll(dir)
	//log.Println(os.TempDir())

	// run chron job
	fmt.Printf("%s Running...\n", time.Now().Format("2006-01-02 15:04:05"))
	ticker := time.NewTicker(time.Minute * 20)
	defer ticker.Stop()
	done := make(chan bool)

	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case _ = <-ticker.C:

			// sleep random
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(300) // 5min
			time.Sleep(time.Duration(n) * time.Second)

			// only from 7:00 till 23:00
			hours, _, _ := time.Now().Clock()
			if hours > 6 && hours < 23 {

				startInfo, starts, status, err := getStarts(dir)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(time.Now().Format("2006-01-02 15:04:05"), starts)

				// set first
				if lastStarts == "" {
					lastStarts = starts
					lastStatus = status
				}

				// if something changes
				if lastStarts != starts || lastStatus != status {
					lastStarts = starts
					lastStatus = status
					fmt.Println("NEW", starts, "\n", status)
					sendMail("ALERT", startInfo+"\n\n"+status)
				}

			} else {
				lastStarts = ""
				lastStatus = ""
			}

		}
	}

}
