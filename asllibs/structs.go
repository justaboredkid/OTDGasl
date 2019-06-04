// Package asllibs is a internal library for parsing ASL using the OTDG
package asllibs

/*
Copyright (C) 2019  Henry Lo

This file is part of OTDGasl.

    OTDGasl is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    OTDGasl is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with OTDGasl.  If not, see <https://www.gnu.org/licenses/>.
*/

// Hand defines hand data
type Hand struct {
	/*
		Palm button orientation is FACING viewer
		Left is towards the pinky and so on

		Between finger is defined by betw, then the initial of the fingers
	*/
	Pinky     bool        `json:"pinky"`
	Ring      bool        `json:"ring"`
	Middle    bool        `json:"middle"`
	Index     bool        `json:"index"`
	Thumb     bool        `json:"thumb"`
	PalmLeft  bool        `json:"palmLeft"`
	PalmRight bool        `json:"palmRight"`
	BackThumb bool        `json:"backThumb"`
	BackRing  bool        `json:"backRing"`
	BetwIM    bool        `json:"betwIM"`
	BetwMR    bool        `json:"betwMR"`
	BetwRP    bool        `json:"betwRP"`
	Angle     Orientation `json:"angle"`
	Motion    string      `json:"motion"`
	Dom       bool        `json:"dom"`
}

// ASLdict is a struct for parsing the ASL JSON files
type ASLdict struct {
	// BUG(justaboredkid): define face more
	ID       string `json:"id"`
	Hand     Hand   `json:"hand"`
	Location string `json:"location"`
	Face     string `json:"face"`
}

// Orientation is a struct for parsing AJAX from sensor
type Orientation struct {
	Alpha int `json:"alpha"`
	Beta  int `json:"beta"`
	Gamma int `json:"gamma"`
}
