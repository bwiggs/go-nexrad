# go-nexrad-geojson: NEXRAD to GeoJSON

![](https://img.shields.io/badge/status-alpha-red.svg?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/jtleniger/go-nexrad-geojson?style=flat-square)](https://goreportcard.com/report/github.com/jtleniger/go-nexrad-geojson)
[![license](https://img.shields.io/github/license/jtleniger/go-nexrad-geojson.svg?style=flat-square)](https://raw.githubusercontent.com/jtleniger/go-nexrad-geojson/master/LICENSE)

Create GeoJSON from NEXRAD data.

## Features

- Create GeoJSON from NEXRAD Level 2 (Archive II Format)
	- Output
		- Polygons for each bin for a given product
		- Single elevation or range of elevations
	- Products 
		- Reflectivity (REF)
		- Velocity (VEL)
		- Spectrum Width (SW)
		- Correlation Coefficient (RHO)
		- Differential Reflectivity (ZDR)
		- Differential Phase Shift (PHI)

## Dependencies

- [PROJ](https://proj.org/) version 6 or higher 
	- On Debian-based systems:
		- proj-bin
    	- proj-data
    	- libproj-dev
- pbzip2 (optional, increases performace when decompressing bzip2)

## Additional Reading

- [NOAA - Introduction to Doppler Radar](http://www.srh.noaa.gov/jetstream/doppler/doppler_intro.html) - Overview of Doppler Radar Technology
- [WSR-88D Govenment Training Course](http://training.weather.gov/wdtd/courses/rac/intro/rda/index.html) - Overview of the WSR-88D Radar and system components.
- [NWS WSR-88D Radar Fundamentals - Slide Deck](https://www.meteor.iastate.edu/classes/mt432/lectures/ISURadarTalk_NWS_2013.pdf) - Kevin Skow National Weather Service, Des Moines, IA
- [AWS S3 NEXRAD Data Browser](https://s3.amazonaws.com/noaa-nexrad-level2/index.html) - Amazon hosted NEXRAD data.
- [GCP Cloud Storage Console - NEXRAD Data Browser](https://console.cloud.google.com/storage/browser/gcp-public-data-nexrad-l2/) - Google hosted NEXRAD data.
