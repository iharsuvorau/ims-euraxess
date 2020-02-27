package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	diff "github.com/sergi/go-diff/diffmatchpatch"
)

func newTestServer() (ts *httptest.Server, err error) {
	mux := http.NewServeMux()
	mux.Handle("/jobs/421010", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/0.html")
	}))
	mux.Handle("/jobs/434505", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/1.html")
	}))
	mux.Handle("/jobs/431934", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/2.html")
	}))
	mux.Handle("/jobs/415416", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/3.html")
	}))
	mux.Handle("/jobs/407546", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/4.html")
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/list.html")
	}))
	return httptest.NewServer(mux), nil
}

func Test_collectOfferLinks(t *testing.T) {
	ts, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	defer ts.Close()
	tsBase := "http://" + ts.Listener.Addr().String()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []offerLink
		wantErr bool
	}{
		{
			name: "A",
			args: args{path: ts.URL},
			want: []offerLink{
				offerLink{
					title: "Post-doctoral researcher in Smart Maintenance using Artificial Intelligence",
					uri:   tsBase + "/jobs/421010",
				},
				offerLink{
					title: "Research Fellow in Surgical Robotics",
					uri:   tsBase + "/jobs/434505",
				},
				offerLink{
					title: "Assistant Professor Dynamic Behaviour of Interactive Materials",
					uri:   tsBase + "/jobs/431934",
				},
				offerLink{
					title: "Tenure-track Assistant Professor in Electrophysiological patient monitoring",
					uri:   tsBase + "/jobs/415416",
				},
				offerLink{
					title: "BOF-77 Post-Doctoral Researcher in Energy Harvesting in Industry 4.0",
					uri:   tsBase + "/jobs/407546",
				},
			},
			wantErr: false,
		},
		{
			name:    "B",
			args:    args{path: "/foobar"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collectOfferLinks(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("collectOfferLinks() error = %v, \nwantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				s1 := fmt.Sprintf("%+v", got)
				s2 := fmt.Sprintf("%+v", tt.want)
				d := diff.New()
				diffs := d.DiffPrettyText(d.DiffMain(s1, s2, false))
				t.Errorf("collectOfferLinks() = %v, \nwant %v\ndiff: %v", got, tt.want, diffs)
			}
		})
	}
}

func Test_collectOffers(t *testing.T) {
	ts, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	defer ts.Close()

	links, err := collectOfferLinks(ts.URL)
	if err != nil {
		t.Error(err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	type args struct {
		links []offerLink
	}
	tests := []struct {
		name    string
		args    args
		want    []Offer
		wantErr bool
	}{
		{
			name:    "A",
			args:    args{links},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectOffers(tt.args.links, logger)
			if len(got) != len(links) {
				t.Errorf("want: %v, got: %v", len(got), len(links))
				return
			}
		})
	}
}

func Benchmark_collectOffersSequential(b *testing.B) {
	ts, err := newTestServer()
	if err != nil {
		b.Error(err)
	}
	defer ts.Close()

	links, err := collectOfferLinks(ts.URL)
	if err != nil {
		b.Errorf("failed to collect links: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := collectOffersSequential(links, logger)
		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_collectOffers(b *testing.B) {
	ts, err := newTestServer()
	if err != nil {
		b.Error(err)
	}
	defer ts.Close()

	links, err := collectOfferLinks(ts.URL)
	if err != nil {
		b.Errorf("failed to collect links: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collectOffers(links, logger)
	}
}
