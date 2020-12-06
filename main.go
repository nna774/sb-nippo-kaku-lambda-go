package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	l "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	// YYYYMMDD is golang lunatic parse sting https://golang.org/pkg/time/#pkg-constants
	YYYYMMDD = "2006/01/02"
	// MD is too.
	MD = "1月2日"
	// TZ is timezone
	TZ = "Asia/Tokyo"
	// Day is 24 hours
	Day = time.Hour * 24
)

var (
	// Project is the Scrapbox project for make Nippo
	Project = os.Getenv("PROJECT")
	// Lastorder is the arn of lastOrder
	Lastorder = os.Getenv("lastOrder")
	// Weekday is Japanese string of weekday
	Weekday = []string{
		"日曜日",
		"月曜日",
		"火曜日",
		"水曜日",
		"木曜日",
		"金曜日",
		"土曜日",
	}
)

// Input is the input of this lambda
type Input struct {
	Date string `json:"date"`
}

// Response is the response of thi lambda
type Response struct {
	Location string `json:"location"`
}

type lastOrderOutput struct {
	Last string `json:"last"`
}

func fail(cause error) (Response, error) {
	return Response{}, cause
}

func redirect(to string) (Response, error) {
	return Response{Location: to}, nil
}

func dateEqual(d1, d2 time.Time) bool {
	return d1.Format(YYYYMMDD) == d2.Format(YYYYMMDD)
}

// MakeNippoHandler makes Nippo redirection
func MakeNippoHandler(ctx context.Context, input Input) (Response, error) {
	loc, err := time.LoadLocation(TZ)
	if err != nil {
		return fail(errors.Wrap(err, "wrong TZ string"))
	}
	theDay := time.Now().In(loc)
	isToday := true
	if input.Date != "" { // with date parameter
		isToday = false
		theDay, err = time.ParseInLocation(YYYYMMDD, input.Date, loc)
		if err != nil {
			return fail(errors.Wrap(err, fmt.Sprintf("%v is not valid date format", input.Date)))
		}
	}

	redirectTo := Project + "/" + url.PathEscape(theDay.Format(YYYYMMDD))
	if isToday {
		// if today, invoke lastorder; it saves the last time invoked
		svc := l.New(session.New())
		payload := []byte(`{"name": "sb-nippo-kaku-lambda-go"}`)
		input := &l.InvokeInput{
			FunctionName: aws.String(Lastorder),
			Payload:      payload,
		}
		res, err := svc.Invoke(input)
		if err != nil {
			return fail(errors.Wrap(err, "lastorder invoke failed"))
		}
		var lo lastOrderOutput
		err = json.Unmarshal(res.Payload, &lo)
		if err != nil {
			return fail(errors.Wrap(err, "lastorder response unmarshal fail"))
		}
		if lo.Last != "" {
			lasttime, err := strconv.ParseInt(lo.Last, 10, 64)
			if err != nil {
				return fail(errors.Wrap(err, "lastorder response parseInt fail"))
			}
			lastDay := time.Unix(lasttime, 0).In(loc)
			if dateEqual(lastDay, theDay) {
				return redirect(redirectTo)
			}
		}
	}
	var body string
	body = body + "\n"

	body = body + "#" + theDay.Format(MD) + " "
	body = body + "#" + theDay.Add(-1 * Day).Format(YYYYMMDD) + " #" + theDay.Add(Day).Format(YYYYMMDD) + "\n"
	body = body + "#" + Weekday[theDay.Weekday()] + " #nippo"
	redirectTo = redirectTo + "?body=" + url.QueryEscape(body)
	return redirect(redirectTo)
}

func main() {
	lambda.Start(MakeNippoHandler)
}
