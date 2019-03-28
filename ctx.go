package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type result struct {
	res []byte
	err error
}

var (
	aLog = newLogger("a")
	bLog = newLogger("b")
	cLog = newLogger("c")
)

// newLogger returns a logger with a name
func newLogger(name string) *log.Logger {
	return log.New(os.Stdout, name+" ", log.Ldate|log.Ltime)
}

// httpReq adds a context and runs a http call
// then parses and returns the body
func httpReq(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err // could be context deadline exceeded
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body) // could be context deadline exceeded
}

func main() {
	c := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cLog.Println("got a call")
		ctx := r.Context()
		done := make(chan bool)
		
		go func() {
			time.Sleep(1 * time.Second) // artificial delay
			done <- true
		}()

		select {
		case <-ctx.Done():
			cLog.Println("timeout")
			fmt.Fprintln(w, ctx.Err()) // http.Error ?
		case <-done:
			cLog.Println("success")
			fmt.Fprintln(w, "hello from c")
		}
	}))
	cLog.Printf("Running on %s\n", c.URL)
	defer c.Close()

	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bLog.Println("got a call")
		ctx := r.Context()
		done := make(chan result)

		go func() {
			bLog.Println("calling c")
			res, err := httpReq(ctx, c.URL)
			if err != nil {
				done <- result{err: fmt.Errorf("%s", err)}
				return
			}
			time.Sleep(1 * time.Second) // artificial delay
			done <- result{res: res}
		}()

		select {
		case <-ctx.Done():
			bLog.Println("timeout")
			fmt.Fprintln(w, ctx.Err())
		case res := <-done:
			if res.err != nil {
				bLog.Println("got an error from c")
				http.Error(w, http.StatusText(500), 500)
				return
			}
			bLog.Println("success")
			w.Write(res.res)
		}
	}))
	bLog.Printf("Running on %s\n", b.URL)
	defer b.Close()

	// a
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	aLog.Println("calling b")
	res, err := httpReq(ctx, b.URL)
	if err != nil {
		aLog.Fatal(err)
	}
	aLog.Println(string(res))
}
