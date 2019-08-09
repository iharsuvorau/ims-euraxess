package main

import (
	"fmt"
	"reflect"
	"testing"

	diff "github.com/sergi/go-diff/diffmatchpatch"
)

func Test_collectOfferLinks(t *testing.T) {
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
			args: args{path: "https://euraxess.ec.europa.eu/jobs/search?keywords=Intelligent%20Materials%20and%20Systems%20Lab"},
			want: []offerLink{
				offerLink{
					title: "Post-doctoral researcher in Smart Maintenance using Artificial Intelligence",
					uri:   "https://euraxess.ec.europa.eu/jobs/421010",
				},
				offerLink{
					title: "Research Fellow in Surgical Robotics",
					uri:   "https://euraxess.ec.europa.eu/jobs/434505",
				},
				offerLink{
					title: "Assistant Professor Dynamic Behaviour of Interactive Materials",
					uri:   "https://euraxess.ec.europa.eu/jobs/431934",
				},
				offerLink{
					title: "Tenure-track Assistant Professor in Electrophysiological patient monitoring",
					uri:   "https://euraxess.ec.europa.eu/jobs/415416",
				},
				offerLink{
					title: "BOF-77 Post-Doctoral Researcher in Energy Harvesting in Industry 4.0",
					uri:   "https://euraxess.ec.europa.eu/jobs/407546",
				},
				offerLink{
					title: "BOF-66 Post-Doctoral Researcher in Wearable Systems for Human Computer Interfacing in Industry 4.0",
					uri:   "https://euraxess.ec.europa.eu/jobs/419593",
				},
			},
			wantErr: false,
		},
		{
			name:    "B",
			args:    args{path: "https://euraxess.ec.europa.eu/jobs/fooobar"},
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

// func Test_collectOffers(t *testing.T) {
// 	type args struct {
// 		links []offerLink
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []offer
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := collectOffers(tt.args.links)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("collectOffers() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("collectOffers() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_collectOffer(t *testing.T) {
	type args struct {
		link offerLink
	}
	tests := []struct {
		name    string
		args    args
		want    offer
		wantErr bool
	}{
		{
			name:    "A",
			args:    args{link: offerLink{title: "Post-doctoral researcher in Smart Maintenance using Artificial Intelligence", uri: "https://euraxess.ec.europa.eu/jobs/421010"}},
			want:    offer{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collectOffer(tt.args.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("collectOffer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Error("must not be nil")
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("collectOffer() = %+v, want %+v", got, tt.want)
			// }
		})
	}
}
