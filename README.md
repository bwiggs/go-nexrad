# NEXRAD Level 2 Product Processing

A go application to process raw archive2 NEXRAD data and generate radar imagery.

![Hurricane Harvey after landfall](ref.png)
Reflectivity Radar for Hurricane Harvey after making landfall from the Corpus Christy Radar site.

## Resources

### NEXRAD

![WSR-88D Photo](http://training.weather.gov/wdtd/courses/rac/intro/graphics/radar.jpg)

> NEXRAD or Nexrad (Next-Generation Radar) is a network of 159 high-resolution S-band Doppler weather radars operated by the National Weather Service (NWS), an agency of the National Oceanic and Atmospheric Administration (NOAA) within the United States Department of Commerce, the Federal Aviation Administration (FAA) within the Department of Transportation, and the U.S. Air Force within the Department of Defense. Its technical name is WSR-88D, which stands for Weather Surveillance Radar, 1988, Doppler.
> 
> NEXRAD detects precipitation and atmospheric movement or wind. It returns data which when processed can be displayed in a mosaic map which shows patterns of precipitation and its movement. The radar system operates in two basic modes, selectable by the operator – a slow-scanning clear-air mode for analyzing air movements when there is little or no activity in the area, and a precipitation mode, with a faster scan for tracking active weather. NEXRAD has an increased emphasis on automation, including the use of algorithms and automated volume scans.

> [NEXRAD: Wikipedia](https://en.wikipedia.org/wiki/NEXRAD)

### Archive 2 (Level 2) Data Format

> Level II data are sometimes referred to as “base data.” Level II data contain the reflectivity, radial velocity, and spectrum width data produced by the WSR-88D. For sites where the Dual Polarization modification has been completed, the following dual polarization moment data are also included: Differential Reflectivity, Correlation Coefficient, and Differential Phase. **They contain the data from all scans of the radar, at 256 data levels, and at the highest spatial resolution of the radar (1ox 1km for reflectivity, 1ox 0.25 km for radial velocity, and 1o x 0.25 km for spectrum width). At lower elevation angles (generally scans at 1.5o or lower), Super Resolution Data are produced. The difference is that Super Resolution has the following spatial resolution (0.5o x 0.25km for reflectivity, 0.5o x 0.25 km for radial velocity, and 0.5o x 0.25 km for spectrum width). In addition, Super Resolution data contain Doppler data out to a range of 300 km. For more information, go to: http://www.roc.noaa.gov/WSR88D/DualPol/DPLevelII.aspx.**

### Glossary

- **RADAR** - RAdio Detection And Ranging
- **NEXRAD**: Next Generation Radar
- **RDA**: Radar Data Acquisition
- **RPG**: Radar Product Generations
- **NetCDF** Network Common Data Form - is a set of software libraries and self-describing, machine-independent data formats that support the creation, access, and sharing of array-oriented scientific data.
- **ICAO** International Civil Aviation Organization

## Links

- [NOAA - Introduction to Doppler Radar](http://www.srh.noaa.gov/jetstream/doppler/doppler_intro.html) - Overview of Doppler Radar Technology
- [WSR-88D Govenment Training Course](http://training.weather.gov/wdtd/courses/rac/intro/rda/index.html) - Overview of the WSR-88D Radar and system components.
- [NWS WSR-88D Radar Fundamentals - Slide Deck](https://www.meteor.iastate.edu/classes/mt432/lectures/ISURadarTalk_NWS_2013.pdf) - Kevin Skow National Weather Service, Des Moines, IA