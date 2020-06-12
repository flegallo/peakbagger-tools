package peakbagger

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	c "peakbagger-tools/pbtools/convert"
	"peakbagger-tools/pbtools/track"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/tkrajina/gpxgo/gpx"
)

// PeakBagger information
type PeakBagger struct {
	Username   string
	Password   string
	ClimberID  string
	HTTPClient *http.Client
}

type aspNetContext struct {
	PageTitle          string
	EventValidation    string
	ViewStateGenerator string
	ViewState          string
}

type peaksXML struct {
	XMLName xml.Name  `xml:"ts"`
	Peaks   []peakXML `xml:"t"`
}

type peakXML struct {
	XMLName   xml.Name `xml:"t"`
	PeakID    string   `xml:"i,attr"`
	Latitude  float64  `xml:"a,attr"`
	Longitude float64  `xml:"o,attr"`
	Name      string   `xml:"n,attr"`
}

const baseURL = "https://peakbagger.com"
const formDataBoundary = "-----------------------------17633381196503435833281039455"

// NewClient creates a new client to interact with PeakBagger website
func NewClient(username string, password string) *PeakBagger {

	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{Jar: cookieJar}

	return &PeakBagger{
		Username:   username,
		Password:   password,
		ClimberID:  "",
		HTTPClient: &httpClient,
	}
}

// Login tries to log in to PeakBagger website
func (pb *PeakBagger) Login() (string, error) {
	page := "Climber/Login.aspx"
	fullURL := fmt.Sprintf("%s/%s", baseURL, page)

	aspNetContext, err := pb.getAspNetContextData(page)
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Add("__EVENTVALIDATION", aspNetContext.EventValidation)
	form.Add("__VIEWSTATEGENERATOR", aspNetContext.ViewStateGenerator)
	form.Add("__VIEWSTATE", aspNetContext.ViewState)
	form.Add("EmailTextBox", pb.Username)
	form.Add("PasswordTextBox", pb.Password)
	form.Add("GoButton", "Log In")

	req, err := http.NewRequest("POST", fullURL, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := pb.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("peakbagger login failed with error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	loginMessage := doc.Find("#MessageBox").First().Text()
	if !strings.Contains(loginMessage, "Successful Login") {
		return "", fmt.Errorf("peakbagger login failed with error: '%s'", loginMessage)
	}

	href, _ := doc.Find("a:contains('My Home Page')").Next().Attr("href")
	climberID, _ := parsePeakbaggerIDFromURL(href, "cid")
	pb.ClimberID = climberID

	return climberID, nil
}

// AddAscent adds an ascent in Peakbagger.com
func (pb *PeakBagger) AddAscent(ascent Ascent) (string, error) {

	ctx, err := pb.uploadGPX(ascent.PeakID, ascent.Gpx)
	if err != nil {
		return "", err
	}

	page := fmt.Sprintf("climber/ascentedit.aspx?pid=%s&cid=%s", ascent.PeakID, pb.ClimberID)
	fullURL := fmt.Sprintf("%s/%s", baseURL, page)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.SetBoundary(formDataBoundary)

	// TODO complete some params and move that to a struct instead
	writer.WriteField("__EVENTVALIDATION", ctx.EventValidation)
	writer.WriteField("__VIEWSTATEGENERATOR", ctx.ViewStateGenerator)
	writer.WriteField("__VIEWSTATE", ctx.ViewState)
	writer.WriteField("PointFt", "0")
	writer.WriteField("PointM", "0")
	writer.WriteField("DateText", ascent.Date.Format("2006-01-02"))
	writer.WriteField("SaveButton", "Save Ascent")
	writer.WriteField("AscentTypeRBL", "S")
	writer.WriteField("JournalText", ascent.TripReport)

	writer.WriteField("StartFt", c.Ftoan(c.ToFeet(ascent.StartElevation)))
	if ascent.NetGain >= 0 {
		writer.WriteField("GainFt", c.Ftoan(c.ToFeet(ascent.NetGain)))
		writer.WriteField("GainM", c.Ftoan(ascent.NetGain))
	}
	if ascent.ExtraGainUp >= 0 {
		writer.WriteField("ExUpFt", c.Ftoan(c.ToFeet(ascent.ExtraGainUp)))
		writer.WriteField("ExUpM", c.Ftoan(ascent.ExtraGainUp))
	}
	if ascent.DistanceUp >= 0 {
		writer.WriteField("UpMi", c.Ftoan(c.ToMiles(ascent.DistanceUp)))
		writer.WriteField("UpKm", c.Ftoan(ascent.DistanceUp/1000))
	}
	if ascent.TimeUp >= 0 {
		d, h, m := c.ToDaysHoursMin(ascent.TimeUp)
		writer.WriteField("UpDay", strconv.Itoa(d))
		writer.WriteField("UpHr", strconv.Itoa(h))
		writer.WriteField("UpMin", strconv.Itoa(m))
	}

	writer.WriteField("EndFt", c.Ftoan(c.ToFeet(ascent.EndElevation)))
	if ascent.NetLoss >= 0 {
		writer.WriteField("LossFt", c.Ftoan(c.ToFeet(ascent.NetLoss)))
		writer.WriteField("LossM", c.Ftoan(ascent.NetLoss))
	}
	if ascent.ExtraLossDown >= 0 {
		writer.WriteField("ExDnFt", c.Ftoan(c.ToFeet(ascent.ExtraLossDown)))
		writer.WriteField("ExDnM", c.Ftoan(ascent.ExtraLossDown))
	}
	if ascent.DistanceDown >= 0 {
		writer.WriteField("DnMi", c.Ftoan(c.ToMiles(ascent.DistanceDown)))
		writer.WriteField("DnKm", c.Ftoan(ascent.DistanceDown/1000))
	}
	if ascent.TimeDown >= 0 {
		d, h, m := c.ToDaysHoursMin(ascent.TimeDown)
		writer.WriteField("DnDay", strconv.Itoa(d))
		writer.WriteField("DnHr", strconv.Itoa(h))
		writer.WriteField("DnMin", strconv.Itoa(m))
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fullURL, body)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+formDataBoundary)

	res, err := pb.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("peakbagger add ascent failed with error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	message := doc.Find("span#SubTitle").Text()
	if message == "" {
		return "", fmt.Errorf("peakbagger add ascent failed with unknown error")
	}
	if !strings.Contains(message, "Saved Successfully") {
		return "", fmt.Errorf("peakbagger add ascent failed with error: '%s'", message)
	}

	return "", nil
}

// DeleteAscent deletes an ascent from peakbagger.com
func (pb *PeakBagger) DeleteAscent(ascentID string) error {
	page := fmt.Sprintf("climber/ascentedit.aspx?aid=%s", ascentID)
	fullURL := fmt.Sprintf("%s/%s", baseURL, page)

	ctx, err := pb.getAspNetContextData(page)
	if err != nil {
		return err
	}

	if strings.Contains(ctx.PageTitle, "Invalid User") {
		return fmt.Errorf("invalid id")
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.SetBoundary(formDataBoundary)

	// TODO complete some params and move that to a struct instead
	writer.WriteField("__EVENTVALIDATION", ctx.EventValidation)
	writer.WriteField("__VIEWSTATEGENERATOR", ctx.ViewStateGenerator)
	writer.WriteField("__VIEWSTATE", ctx.ViewState)
	writer.WriteField("DeleteButton", "Delete Ascent") // yes that's what differentiate a delete from an add...

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fullURL, body)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+formDataBoundary)

	res, err := pb.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("peakbagger delete ascent failed with error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	message := doc.Find("span#SubTitle").Text()
	if message == "" {
		return fmt.Errorf("peakbagger delete ascent failed with unknown error")
	}
	if !strings.Contains(message, "Ascent Deleted") {
		return fmt.Errorf("peakbagger delete ascent failed with error: '%s'", message)
	}

	return nil

}

// FindPeaks find a list of peaks near the given location
func (pb *PeakBagger) FindPeaks(bounds *track.Bounds) ([]Peak, error) {
	url := fmt.Sprintf("%s/Async/PLLBB.aspx?miny=%f&maxy=%f&minx=%f&maxx=%f",
		baseURL,
		bounds.MinLat,
		bounds.MaxLat,
		bounds.MinLng,
		bounds.MaxLng,
	)

	resp, err := pb.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't search peaks in peakbaggers: '%s'", err)
	}

	byteValue, _ := ioutil.ReadAll(resp.Body)

	var peaks peaksXML
	err = xml.Unmarshal(byteValue, &peaks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peaks result xml: '%s'", err)
	}

	results := make([]Peak, len(peaks.Peaks))
	for i, p := range peaks.Peaks {
		results[i] = Peak{
			PeakID:    p.PeakID,
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
			Name:      p.Name,
		}
	}

	return results, nil
}

// ListAscents list ascents for the logged user
func (pb *PeakBagger) ListAscents() (ClimberAscents, error) {
	res, err := pb.HTTPClient.Get(fmt.Sprintf("%s/climber/ClimbListC.aspx?cid=%s&u=m&sort=AscentDate&y=9999", baseURL, pb.ClimberID))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to load climber ascents page: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	ascents := []AscentSummary{}
	var parseErr error
	doc.Find("table.gray > tbody > tr").Each(func(index int, sel *goquery.Selection) {
		tds := sel.Find("td")
		if tds.Length() == 10 {
			urlName := tds.Children().First()
			linkDate := tds.Next().Children().First()

			peakURL, pExists := urlName.Attr("href")
			name := urlName.Text()
			ascentURL, aExists := linkDate.Attr("href")
			dateText := linkDate.Text()

			elevationS := tds.Eq(3).Text()
			location := tds.Eq(4).Text()

			elevation, err := strconv.ParseFloat(elevationS, 64)

			if !pExists || !aExists || err != nil {
				parseErr = errors.New("something failed while trying to parse ascent")
				return
			}

			date, err := time.Parse("2006-01-02", dateText)
			peakID, pExists := parsePeakbaggerIDFromURL(peakURL, "pid")
			ascentID, aExists := parsePeakbaggerIDFromURL(ascentURL, "aid")

			if !pExists || !aExists || err != nil {
				parseErr = errors.New("something failed while trying to parse ascent")
				return
			}

			ascents = append(ascents, AscentSummary{
				AscentID:  ascentID,
				PeakID:    peakID,
				PeakName:  name,
				Date:      &date,
				Elevation: elevation,
				Location:  location,
			})
		}
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return ascents, nil
}

func parsePeakbaggerIDFromURL(url string, id string) (string, bool) {
	split := strings.Split(url, id+"=")
	if len(split) != 2 {
		return "", false
	}

	res := split[1]
	if strings.Contains(res, "&") {
		res = string(res[0:strings.Index(res, "&")])
	}

	return res, true
}

func (pb *PeakBagger) uploadGPX(peakID string, g *gpx.GPX) (*aspNetContext, error) {

	page := fmt.Sprintf("climber/ascentedit.aspx?pid=%s&cid=%s", peakID, pb.ClimberID)
	fullURL := fmt.Sprintf("%s/%s", baseURL, page)

	ctx, err := pb.getAspNetContextData(page)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.SetBoundary(formDataBoundary)

	writer.WriteField("__EVENTVALIDATION", ctx.EventValidation)
	writer.WriteField("__VIEWSTATEGENERATOR", ctx.ViewStateGenerator)
	writer.WriteField("__VIEWSTATE", ctx.ViewState)
	writer.WriteField("GPXPreview", "Preview")

	xmlBytes, err := g.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: false})
	if err != nil {
		return nil, err
	}
	part, err := writer.CreateFormFile("GPXUpload", "upload.gpx")
	if err != nil {
		return nil, err
	}
	part.Write(xmlBytes)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fullURL, body)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+formDataBoundary)

	res, err := pb.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("peakbagger upload gpx failed with error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	eventValidation, _ := doc.Find("input[name='__EVENTVALIDATION']").Attr("value")
	viewStateGen, _ := doc.Find("input[name='__VIEWSTATEGENERATOR']").Attr("value")
	viewState, _ := doc.Find("input[name='__VIEWSTATE']").Attr("value")

	return &aspNetContext{
		EventValidation:    eventValidation,
		ViewStateGenerator: viewStateGen,
		ViewState:          viewState,
	}, nil
}

func (pb *PeakBagger) getAspNetContextData(path string) (*aspNetContext, error) {
	res, err := pb.HTTPClient.Get(fmt.Sprintf("%s/%s", baseURL, path))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to load page '%s': %d %s", path, res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	pageTitle := doc.Find("span#PageTitle > h1").Text()
	eventValidation, _ := doc.Find("input[name='__EVENTVALIDATION']").Attr("value")
	viewStateGen, _ := doc.Find("input[name='__VIEWSTATEGENERATOR']").Attr("value")
	viewState, _ := doc.Find("input[name='__VIEWSTATE']").Attr("value")

	return &aspNetContext{
		PageTitle:          pageTitle,
		EventValidation:    eventValidation,
		ViewStateGenerator: viewStateGen,
		ViewState:          viewState,
	}, nil
}
