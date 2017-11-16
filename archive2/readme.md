# The Archive 2 File Format

> The following is an excerpt from the `ICD for the Archive II User` document.

### 7.3.6.1.1 LDM Raw Data File Format

To exploit the Archive II data the end user must develop a program to extract and decompress the bz2 data stored in the LDM raw data file. Once decompressed each message requires 2432 bytes of storage with the exception of Message Type 31 (Digital Radar Data Generic Format) which is variable length.

#### 1. Volume Header Record
The first record in the file is the Volume Header Record, a 24-byte record. This record will contain the volume number along with a date and time field.

#### 2. LDM Compressed Record

This second record is bzip2 compressed. It consists of Metadata message types 15, 13, 18, 3, 5, and 2. See section 7.3.5. This is all related to radar site status.


#### 3-N. LDM Compressed Record

A variable size record that is bzip2 compressed. It consists of 120 radial data messages (type 1 or 31) plus 0 or more RDA Status messages (type 2). 

The last message will have a radial status signaling “end of elevation” or “end of volume”. See paragraph 7.3.4. Repeat (LDM Compressed Record) Or End of File (for end of volume data) 

### Message Types

| Message Type | Description |
|--------------|-------------|
|Message 2|RDA Status Data, contains the state of operational functions|
|Message 3|RDA Performance/Maintenance Data|
|Message 5|RDA Volume Coverage Pattern|
|Message 13|RDA Clutter Filter Bypass Map|
|Message 15|RDA Clutter Map Data|
|Message 18|RDA Adaptation Data|
|Message 31|Digital Radar Data Generic Format|

## Resources

- [List of ICDs (Interface Control Documents)](https://www.roc.noaa.gov/WSR88D/BuildInfo/Files.aspx) - Consider these the documentation specs for how data is transported and processed from the Radar unit into radar products.
- [ICD for RDA/RPG](https://www.roc.noaa.gov/wsr88d/PublicDocs/ICDs/RDA_RPG_2620002P.pdf) - The spec on the file format and field values on the archive 2 data format.
- [INTERFACE CONTROL DOCUMENT FOR THE ARCHIVE II/USER](https://www.roc.noaa.gov/wsr88d/PublicDocs/ICDs/2620010E.pdf) - This document describes how the radar unit transfers data to the RPG (radar product generator - the thing that makes radar images). It also contains a high level overview of the Archive2 data format.  