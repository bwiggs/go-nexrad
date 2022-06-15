
function DMS_to_Degrees(d,m,s,dir) {
	var deg = parseFloat(Math.abs(d)) + parseFloat(Math.abs(m)/60) + parseFloat(Math.abs(s)/3600);
	if (dir == 'W' || dir == 'S') { deg = parseFloat(-1*deg); }
	if (d == '' && m == '' && s == '') { deg = ''; }
	return comma2point(deg);
}

function Degrees_to_DMM(deg,type,spacer,minutemark) {
	if (!deg.toString().match(/[0-9]/)) { return ''; }
	if (!spacer) { spacer = ''; }
	if (!minutemark) { minutemark = ''; }
	if (type == 'lat') {	
		if (parseFloat(deg) < 0) { var dir = 'S'; } else { var dir = 'N'; }
	} else {
		if (parseFloat(deg) < 0) { var dir = 'W'; } else { var dir = 'E'; }
	}
	var d = Math.floor(Math.abs(parseFloat(deg)));
	var m = 60 * (Math.abs(parseFloat(deg)) - parseFloat(d))
	m = Math.round(1000000 * m) / 1000000;
	if (type == 'lon') {
		if (d < 10) { d = '00'+d; } else if (d < 100) { d = '0'+d; }
	} else {
		if (d < 10) { d = '0'+d; }
	}
	if (parseFloat(m) == Math.floor(parseFloat(m))) { m = m + '.0'; }
	return dir + d + String.fromCharCode(176) + spacer + comma2point(m) + minutemark;
}

function Degrees_to_DMS(deg,type,spacer) {
	if (!deg.toString().match(/[0-9]/)) { return ''; }
	if (!spacer) { spacer = ''; }
	if (type == 'lat') {	
		if (parseFloat(deg) < 0) { var dir = 'S'; } else { var dir = 'N'; }
	} else {
		if (parseFloat(deg) < 0) { var dir = 'W'; } else { var dir = 'E'; }
	}
	var d = Math.floor(Math.abs(parseFloat(deg)));
	var mmm = 60 * (Math.abs(parseFloat(deg)) - parseFloat(d))
	mmm = Math.round(1000000 * mmm) / 1000000;
	var m = Math.floor(parseFloat(mmm));
	var s = 60 * (parseFloat(mmm) - parseFloat(m))
	s = Math.round(1000 * s) / 1000;
	return dir + d + String.fromCharCode(176) + spacer + m + '\'' + spacer + comma2point(s) + '"';
}

function deg2rad (deg) {
	return (parseFloat(comma2point(deg)) * 3.14159265358979/180);
}
function rad2deg (radians) {
	return (Math.round(10000000 * parseFloat(radians) * 180/3.14159265358979) / 10000000);
}
function comma2point (number) {
	number = number+''; // force number into a string context
	return (number.replace(/,/g,'.'));
}

function parseCoordinate(coordinate,type,format,spaced) {
	coordinate = coordinate.toString();
	coordinate = coordinate.replace(/(^\s+|\s+$)/g,''); // remove white space
	var neg = 0; if (coordinate.match(/(^-|[WS])/i)) { neg = 1; }
	if (coordinate.match(/[EW]/i) && !type) { type = 'lon'; }
	if (coordinate.match(/[NS]/i) && !type) { type = 'lat'; }
	coordinate = coordinate.replace(/[NESW\-]/gi,' ');
	if (!coordinate.match(/[0-9]/i)) {
		return '';
	}
	parts = coordinate.match(/([0-9\.\-]+)[^0-9\.]*([0-9\.]+)?[^0-9\.]*([0-9\.]+)?/);
	if (!parts || parts[1] == null) {
		return '';
	} else {
		n = parseFloat(parts[1]);
		if (parts[2]) { n = n + parseFloat(parts[2])/60; }
		if (parts[3]) { n = n + parseFloat(parts[3])/3600; }
		if (neg && n >= 0) { n = 0 - n; }
		if (format == 'dmm') {
			if (spaced) {
				n = Degrees_to_DMM(n,type,' ');
			} else {
				n = Degrees_to_DMM(n,type);
			}
		} else if (format == 'dms') {
			if (spaced) {
				n = Degrees_to_DMS(n,type,' ');
			} else {
				n = Degrees_to_DMS(n,type,'');
			}
		} else {
			n = Math.round(100000000000 * n) / 100000000000;
			if (n == Math.floor(n)) { n = n + '.0'; }
		}
		return comma2point(n);
	}
}
var parseDistance_units = '';
function parseDistance(distance_text) { // returns meters
	var meters = parseFloat(distance_text.replace(/^.*?([0-9]+\.?[0-9]*|\.[0-9]+).*$/,"$1"));
	if (distance_text.match(/mi/i)) { meters *= 1609.344; parseDistance_units = 'us'; }
	else if (distance_text.match(/(\b|\d)(m\b|meter)/i)) { meters *= 1; }
	else if (distance_text.match(/(\b|\d)(naut|n\.?m|kn)/i)) { meters *= 1852; parseDistance_units = 'nautical'; }
	else if (distance_text.match(/((\b|\d)fe*t|')/i)) { meters *= 0.3048; parseDistance_units = 'us'; }
	else if (distance_text.match(/(\b|\d)(yd|yard)/i)) { meters *= 0.9144; parseDistance_units = 'us'; }
	else { meters *= 1000; } // default is kilometers
	return (meters);
}
function parseBearing(bearing_text) { // returns degrees
	var degrees;
	if (bearing_text.toUpperCase().match(/[NS].*[0-9].*[EW]/i)) {
		parts = bearing_text.toUpperCase().match(/([NS])(.*[0-9].*)([EW])/);
		degrees = parts[2];
		if (parts[1] == 'N' && parts[3] == 'E') { degrees = 0 + parseFloat(parseCoordinate(degrees)); }
		else if (parts[1] == 'N' && parts[3] == 'W') { degrees = 360 - parseFloat(parseCoordinate(degrees)); }
		else if (parts[1] == 'S' && parts[3] == 'E') { degrees = 180 - parseFloat(parseCoordinate(degrees)); }
		else if (parts[1] == 'S' && parts[3] == 'W') { degrees = 180 + parseFloat(parseCoordinate(degrees)); }
	} else {
		degrees = parseFloat(parseCoordinate(bearing_text.replace(/[NSEW]/gi,' ')));
	}
	return degrees;
}

function Haversine_Distance(lat1,lon1,lat2,lon2,us) {
	// http://www.movable-type.co.uk/scripts/LatLong.html
	if (Math.abs(parseFloat(lat1)) > 90 || Math.abs(parseFloat(lon1)) > 180 || Math.abs(parseFloat(lat2)) > 90 || Math.abs(parseFloat(lon2)) > 180) { return 'n/a'; }
	lat1 = deg2rad(lat1); lon1 = deg2rad(lon1);
	lat2 = deg2rad(lat2); lon2 = deg2rad(lon2);
	var dlat = lat2-lat1; // delta
	var dlon = lon2-lon1; // delta
	var alat = (lat1+lat2)/2; // average
	var re = 6378137; // equatorial radius
	var rp = 6356752; // polar radius
	var r45 = re * Math.sqrt( (1 + ( (rp*rp-re*re)/(re*re) ) * (Math.sin(45)*Math.sin(45)) ) ) // radius at 45 degrees; from http://www.newton.dep.anl.gov/askasci/gen99/gen99915.htm
	var a = ( Math.sin(dlat/2) * Math.sin(dlat/2) ) + ( Math.cos(lat1) * Math.cos(lat2) * Math.sin(dlon/2) * Math.sin(dlon/2) );
	var c = 2 * Math.atan( Math.sqrt(a)/Math.sqrt(1-a) );
	var d_ellipse = r45 * c;
	if (us) {
		var dist = d_ellipse / 1609.344;
		if (dist < 1) {
			return (Math.round(5280 * 1 * dist) / 1) + ' ft';
		} else {
			return (Math.round(100 * dist) / 100) + ' mi';
		}
	} else {
		var dist = d_ellipse / 1000;
		if (dist < 1) {
			return (Math.round(1000 * 1 * dist) / 1) + ' m';
		} else {
			return (Math.round(100 * dist) / 100) + ' km';
		}
	}
}

function Vincenty_Distance(lat1,lon1,lat2,lon2,us,meters_only) {
	// http://www.movable-type.co.uk/scripts/LatLongVincenty.html
	if (Math.abs(parseFloat(lat1)) > 90 || Math.abs(parseFloat(lon1)) > 180 || Math.abs(parseFloat(lat2)) > 90 || Math.abs(parseFloat(lon2)) > 180) { return 'n/a'; }
	if (lat1 == lat2 && lon1 == lon2) { return '0'; }
	
	lat1 = deg2rad(lat1); lon1 = deg2rad(lon1);
	lat2 = deg2rad(lat2); lon2 = deg2rad(lon2);

	var a = 6378137, b = 6356752.314245, f = 1/298.257223563;
	var L = lon2 - lon1;
	var U1 = Math.atan((1-f) * Math.tan(lat1));
	var U2 = Math.atan((1-f) * Math.tan(lat2));
	var sinU1 = Math.sin(U1), cosU1 = Math.cos(U1);
	var sinU2 = Math.sin(U2), cosU2 = Math.cos(U2);
	var lambda = L, lambdaP = 2*Math.PI;
	var iterLimit = 50;
	while (Math.abs(lambda-lambdaP) > 1e-12 && --iterLimit > 0) {
		var sinLambda = Math.sin(lambda), cosLambda = Math.cos(lambda);
		var sinSigma = Math.sqrt((cosU2*sinLambda) * (cosU2*sinLambda) + 
		  (cosU1*sinU2-sinU1*cosU2*cosLambda) * (cosU1*sinU2-sinU1*cosU2*cosLambda));
		var cosSigma = sinU1*sinU2 + cosU1*cosU2*cosLambda;
		var sigma = Math.atan2(sinSigma, cosSigma);
		var alpha = Math.asin(cosU1 * cosU2 * sinLambda / sinSigma);
		var cosSqAlpha = Math.cos(alpha) * Math.cos(alpha);
		var cos2SigmaM = (!cosSqAlpha) ? 0 : cosSigma - 2*sinU1*sinU2/cosSqAlpha;
		var C = f/16*cosSqAlpha*(4+f*(4-3*cosSqAlpha));
		lambdaP = lambda;
		lambda = L + (1-C) * f * Math.sin(alpha) * (sigma + C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)));
	}
	if (iterLimit==0) { return (NaN); }  // formula failed to converge
	var uSq = cosSqAlpha*(a*a-b*b)/(b*b);
	var A = 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)));
	var B = uSq/1024 * (256+uSq*(-128+uSq*(74-47*uSq)));
	var deltaSigma = B*sinSigma*(cos2SigmaM+B/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM) - B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)));
	var s = b*A*(sigma-deltaSigma); var m = s;
	if (meters_only) {
		return m;
	} else if (us==2) { // nautical
		var nm = m/1852; var feet = m*3.28084;
		if (feet < 1000) {
			if (feet < 1000) {
				return (Math.round(10*feet) / 10) + ' ft';
			} else {
				return (Math.round(feet) / 1) + ' ft';
			}
		} else {
			return (Math.round(1000*nm) / 1000) + ' NM';
		}
	} else if (us) { // U.S.
		var miles = m/1609.344; var feet = m*3.28084;
		if (feet < 1000) {
			if (feet < 1000) {
				return (Math.round(10*feet) / 10) + ' ft';
			} else {
				return (Math.round(feet) / 1) + ' ft';
			}
		} else {
			return (Math.round(1000*miles) / 1000) + ' mi';
		}
	} else { // metric
		var km = m/1000;
		if (km < 1) {
			if (m < 1000) {
				return (Math.round(10*m) / 10) + ' m';
			} else {
				return (Math.round(m) / 1) + ' m';
			}
		} else {
			return (Math.round(1000*km) / 1000) + ' km';
		}
	}
}

function Bearing(lat1,lon1,lat2,lon2,radians) { // input is in degrees, output is your choice
	// http://www.movable-type.co.uk/scripts/LatLong.html
	if (Math.abs(parseFloat(lat1)) > 90 || Math.abs(parseFloat(lon1)) > 180 || Math.abs(parseFloat(lat2)) > 90 || Math.abs(parseFloat(lon2)) > 180) { return 'n/a'; }
	lat1 = deg2rad(lat1); lon1 = deg2rad(lon1);
	lat2 = deg2rad(lat2); lon2 = deg2rad(lon2);
	var dlat = lat2-lat1; // delta
	var dlon = lon2-lon1; // delta
	var bearing = Math.atan2( (Math.sin(dlon)*Math.cos(lat2)) , (Math.cos(lat1)*Math.sin(lat2) - Math.sin(lat1)*Math.cos(lat2)*Math.cos(dlon)) );
	
	if (radians) {
		return (bearing);
	} else {
		bearing = rad2deg(bearing);
		if (bearing < 0) { bearing += 360; }
		return (Math.round(1000 * bearing) / 1000) + String.fromCharCode(176);
	}
}

function Point_At_Distance_And_Bearing(start_lat,start_lon,distance_text,bearing) { // input is in degrees, km, degrees
	// http://www.fcc.gov/mb/audio/bickel/sprong.html
	var ending_point = []; // output
	var earth_radius = 6378137; // equatorial radius
	// var earth_radius = 6356752; // polar radius
	// var earth_radius = 6371000; // typical radius
	var start_lat_rad = deg2rad(parseCoordinate(start_lat));
	var start_lon_rad = deg2rad(parseCoordinate(start_lon));
	var distance = parseDistance(distance_text);
	
	bearing = parseBearing(bearing);
	if (Math.abs(bearing) >= 360) { bearing = bearing % 360; }
	bearing = (bearing < 0) ? bearing+360 : bearing;
	var isig = (bearing <= 180) ? 1 : 0; // western half of circle = 0, eastern half = 1
	var a = 360-bearing; // this subroutine measures angles COUNTER-clockwise, so +3 becomes +357
	a = deg2rad(a); var bb = (Math.PI/2)-start_lat_rad; var cc = distance/earth_radius;
	var sin_bb = Math.sin(bb); var cos_bb = Math.cos(bb); var cos_cc = Math.cos(cc);
	var cos_aa = cos_bb*cos_cc+(sin_bb*Math.sin(cc)*Math.cos(a));
	if (cos_aa <= -1) { cos_aa = -1; } if (cos_aa >= 1) { cos_aa = 1; }
	var aa = (cos_aa.toFixed(15) == 1) ? 0 : Math.acos(cos_aa);
	var cos_c = (cos_cc-(cos_aa*cos_bb))/(Math.sin(aa)*sin_bb);
	if (cos_c <= -1) { cos_c = -1; } if (cos_c >= 1) { cos_c = 1; }
	var c = (cos_c.toFixed(15) == 1) ? 0 : Math.acos(cos_c);
	var end_lat_rad = (Math.PI/2)-aa;
	var end_lon_rad = start_lon_rad-c;
	if (isig == 1) { end_lon_rad = start_lon_rad + c; }
	if (end_lon_rad > Math.PI) { end_lon_rad = end_lon_rad - (2*Math.PI); }
	if (end_lon_rad < (0-Math.PI)) { end_lon_rad = end_lon_rad + (2*Math.PI); }
	ending_point[0] = rad2deg(end_lat_rad); ending_point[1] = rad2deg(end_lon_rad);
	// Use proportional error to adjust things due to oblate Earth; I'm still not entirely sure how/why this works:
	for (i=0; i<5; i++) {
		var vincenty = Vincenty_Distance(start_lat,start_lon,ending_point[0],ending_point[1],false,true);
		if (Math.abs(start_lon-ending_point[1]) > 180) {
			 // something went haywire
		} else {
			var error = (vincenty != 0) ? distance/vincenty : 1;
			var dlat = ending_point[0]-parseFloat(start_lat); var dlon = ending_point[1]-parseFloat(start_lon);
			ending_point[0] = parseFloat(start_lat)+(dlat*error); ending_point[1] = parseFloat(start_lon)+(dlon*error);
		}
	}
	return (ending_point);
}

function Point_At_Distance_And_Bearing2(start_lat,start_lon,distance_text,bearing) { // input is in degrees, km, degrees
	// http://www.movable-type.co.uk/scripts/latlong.html
	var earth_radius = 6371000; // "average" radius
	var distance = parseDistance(distance_text);
	bearing = deg2rad(parseBearing(bearing));
	var start_lat_rad = deg2rad(parseCoordinate(start_lat));
	var start_lon_rad = deg2rad(parseCoordinate(start_lon));
	var ending_point = []; // output
	var arc = distance/earth_radius;
	var end_lat_rad = Math.asin( Math.sin(start_lat_rad)*Math.cos(arc) + Math.cos(start_lat_rad)*Math.sin(arc)*Math.cos(bearing) );
	var end_lon_rad = start_lon_rad + Math.atan2( Math.sin(bearing)*Math.sin(arc)*Math.cos(start_lat_rad),Math.cos(arc)-Math.sin(start_lat_rad)*Math.sin(end_lat_rad));
	end_lon_rad = (end_lon_rad+Math.PI)%(2*Math.PI) - Math.PI; // normalise to -180...+180
	ending_point[0] = rad2deg(end_lat_rad); ending_point[1] = rad2deg(end_lon_rad);

	return (ending_point);
}


function Combine_Coordinates(lat,lon) {
	if (lat && lon) {
		var coords = lat + "," + lon;
	} else {
		var coords = '';
	}
	return coords;
}

function Separate_Coordinates(lat_box,lon_box) {
	var coords = '';
	if (lat_box.value.match(/\d.*[,\t].*\d/) && !lat_box.value.match(/[,\t].*\1/) && !lon_box.value.match(/\d/)) {
		coords = lat_box.value;
	} else if (lon_box.value.match(/\d.*[,\t].*\d/) && !lon_box.value.match(/[,\t].*\1/) && !lat_box.value.match(/\d/)) {
		coords = lon_box.value;
	}
	if (coords.match(/\d/)) {
		var coord1 = coords.replace(/^\s*([^,\t]+)\s*[,\t].*/,'$1');
		var coord2 = coords.replace(/^.*?[,\t]\s*([^,\t]+)\s*$/,'$1');
		if (coord1.match(/[EW]/i) && coord2.match(/[NS]/i)) {
			lat_box.value = coord2; lon_box.value = coord1;
		} else {
			lat_box.value = coord1; lon_box.value = coord2;
		}
	}
}

function Calculate_Distance_Form() {
	var coords1 = document.distance.distance_coords1.value.split(","); var lat1 = parseCoordinate(coords1[0]); var lon1 = parseCoordinate(coords1[1]);
	var coords2 = document.distance.distance_coords2.value.split(","); var lat2 = parseCoordinate(coords2[0]); var lon2 = parseCoordinate(coords2[1]);
	if (coords1.length > 1 && coords2.length > 1) {
		document.distance.distance_lat1.value = lat1;
		document.distance.distance_lon1.value = lon1;
		document.distance.distance_lat2.value = lat2;
		document.distance.distance_lon2.value = lon2;
		document.distance.distance_metric.value = comma2point(Vincenty_Distance(lat1,lon1,lat2,lon2,0));
		document.distance.distance_us.value = comma2point(Vincenty_Distance(lat1,lon1,lat2,lon2,1));
		document.distance.distance_nautical.value = comma2point(Vincenty_Distance(lat1,lon1,lat2,lon2,2));
		document.distance.distance_bearing.value = comma2point(Bearing(lat1,lon1,lat2,lon2));
	}
}

function Prepare_Distance_Map() {
	document.distance_map.lat1.value = document.distance.distance_lat1.value;
	document.distance_map.lon1.value = document.distance.distance_lon1.value;
	document.distance_map.lat2.value = document.distance.distance_lat2.value;
	document.distance_map.lon2.value = document.distance.distance_lon2.value;
}

function Convert_Coordinates(f) { // f is for "form"
	var lat = f.coordinates_lat.value;
	var lon = f.coordinates_lon.value;
	var spaced = f.coordinates_space.checked;
	f.coordinates_lat_ddd.value = parseCoordinate(lat,'lat','ddd',spaced);
	f.coordinates_lon_ddd.value = parseCoordinate(lon,'lon','ddd',spaced);
	f.coordinates_lat_dmm.value = parseCoordinate(lat,'lat','dmm',spaced);
	f.coordinates_lon_dmm.value = parseCoordinate(lon,'lon','dmm',spaced);
	f.coordinates_lat_dms.value = parseCoordinate(lat,'lat','dms',spaced);
	f.coordinates_lon_dms.value = parseCoordinate(lon,'lon','dms',spaced);
	f.coordinates_pair_ddd.value = f.coordinates_lat_ddd.value+', '+f.coordinates_lon_ddd.value;
	f.coordinates_pair_dmm.value = f.coordinates_lat_dmm.value+', '+f.coordinates_lon_dmm.value;
	f.coordinates_pair_dms.value = f.coordinates_lat_dms.value+', '+f.coordinates_lon_dms.value;
}

function Calculate_Address_Distance_Form() {
	addresses_to_lookup = new Array;
	address_coordinates = new Array;
	addresses_to_lookup[0] = document.distance_address.distance_address_location1.value;
	addresses_to_lookup[1] = document.distance_address.distance_address_location2.value;
	address_lookup_counter = 0;
	var key = '';
	if (document.distance_address.google_api_key.value.match(/\w/)) {
		key = document.distance_address.google_api_key.value.replace(/(^\s+|\s+$)/g,'');
		if (key && !document.distance_address_map.google_api_key.value) {
			document.distance_address_map.google_api_key.value = key;
		}
	}
	if (self.api_key_loaded) {
		GoogleGeocode();
	} else {
		alert('You must enter a valid Google Geocoding API Key to use this form.');
		$('distance_address:google_api_key').className = 'slider_open';
		document.distance_address.google_api_key.focus();
		return false;
	}
}

function Reload_Page_With_API_Key() {
	if (document.distance_address.google_api_key.value.match(/\w/)) {
		key = document.distance_address.google_api_key.value.replace(/(^\s+|\s+$)/g,'');
		if (key) {
			http_post(
				'./calculators#distance_address',{
					distance_address_location1:document.distance_address.distance_address_location1.value,
					distance_address_location2:document.distance_address.distance_address_location2.value,
					google_api_key:key
				},'post'
			);
		}
	}
}

function GoogleGeocode() {
	if (!self.google_api_key) { google_api_key = ''; }
	var loc = addresses_to_lookup[address_lookup_counter];
	var loc_as_coords = check_for_coordinates(loc);
	if (document.images) {
		var image_url = 'calculators-log.png?loc='+encodeURIComponent(loc);
		var im = new Image(); im.src = image_url; // log it
	}

	if (loc_as_coords) {
		googleCallback(loc_as_coords,'coordinates'); // don't actually send it to Google, just go straight to the callback
	} else {
		var gc = new google.maps.Geocoder();
		gc.geocode({'address':loc}, googleCallback); // go to the callback function via Google
	}
}
function googleCallback(results,status) {
//alert('googleCallback status = '+status);
//var x = 'results'; var x0 = eval(x); msg = "Contents of '"+x+"':\n"; for (var x1 in x0) { if (typeof(x0[x1]) == 'object') { for (var x2 in x0[x1]) { if (typeof(x0[x1][x2]) == 'object') { for (var x3 in x0[x1][x2]) { if (typeof(x0[x1][x2][x3]) == 'object') { for (var x4 in x0[x1][x2][x3]) { if (typeof(x0[x1][x2][x3][x4]) == 'object') { for (var x5 in x0[x1][x2][x3][x4]) { if (typeof(x0[x1][x2][x3][x4][x5]) == 'object') { for (var x6 in x0[x1][x2][x3][x4][x5]) { if (typeof(x0[x1][x2][x3][x4][x5][x6]) == 'object') { msg += '// '+x+'['+x1+']['+x2+']['+x3+']['+x4+']['+x5+']['+x6+'] is an object.'+"\n"; } else { msg += x+'['+x1+']['+x2+']['+x3+']['+x4+']['+x5+']['+x6+'] = '+x0[x1][x2][x3][x4][x5][x6]+"\n"; } } } else { msg += x+'['+x1+']['+x2+']['+x3+']['+x4+']['+x5+'] = '+x0[x1][x2][x3][x4][x5]+"\n"; } } } else { msg += x+'['+x1+']['+x2+']['+x3+']['+x4+'] = '+x0[x1][x2][x3][x4]+"\n"; } } } else { msg += x+'['+x1+']['+x2+']['+x3+'] = '+x0[x1][x2][x3]+"\n"; } } } else { msg += x+'['+x1+']['+x2+'] = '+x0[x1][x2]+"\n"; } } } else { msg += x+'['+x1+'] = '+x0[x1]+"\n"; } } alert (msg);
	var coords = [];
	if (status == 'coordinates') {
		coords = results;
	} else if (status && status == google.maps.GeocoderStatus.OK && results && results[0] && results[0].geometry && results[0].geometry.location) {

		var coords = results[0].geometry.location;
		coords[0] = parseFloat(results[0].geometry.location.lat().toFixed(7));
		coords[1] = parseFloat(results[0].geometry.location.lng().toFixed(7));

	} else {
		//alert ('The google geocoder did not return a valid result (status = "'+status+'").');
	}
	address_coordinates.push(coords);
	address_lookup_counter += 1;
	if (address_lookup_counter < addresses_to_lookup.length) {
		GoogleGeocode();
	} else {
		Calculate_Address_Distance_Form2();
	}
}

function Calculate_Address_Distance_Form2() {
	var lat1 = address_coordinates[0][0]; var lon1 = address_coordinates[0][1]; var lat2 = address_coordinates[1][0]; var lon2 = address_coordinates[1][1];
	document.distance_address.distance_address_lat1.value = lat1;
	document.distance_address.distance_address_lon1.value = lon1;
	document.distance_address.distance_address_lat2.value = lat2;
	document.distance_address.distance_address_lon2.value = lon2;
	document.distance_address.distance_address_metric.value = comma2point(Vincenty_Distance(lat1,lon1,lat2,lon2,0));
	document.distance_address.distance_address_us.value = comma2point(Vincenty_Distance(lat1,lon1,lat2,lon2,1));
	document.distance_address.distance_address_nautical.value = comma2point(Vincenty_Distance(lat1,lon1,lat2,lon2,2));
	document.distance_address.distance_address_bearing.value = comma2point(Bearing(lat1,lon1,lat2,lon2));
}
function Prepare_Address_Distance_Map() {
	document.distance_address_map.lat1.value = document.distance_address.distance_address_lat1.value;
	document.distance_address_map.lon1.value = document.distance_address.distance_address_lon1.value;
	document.distance_address_map.lat2.value = document.distance_address.distance_address_lat2.value;
	document.distance_address_map.lon2.value = document.distance_address.distance_address_lon2.value;
	document.distance_address_map.name1.value = document.distance_address.distance_address_location1.value;
	document.distance_address_map.name2.value = document.distance_address.distance_address_location2.value;
	document.distance_address_map.desc1.value = document.distance_address.distance_address_lat1.value+', '+document.distance_address.distance_address_lon1.value;
	document.distance_address_map.desc2.value = document.distance_address.distance_address_lat2.value+', '+document.distance_address.distance_address_lon2.value;
}

function Calculate_Range_Rings_Form() {
	if (document.range_rings.range_rings_coords.value) {
		if (document.range_rings.range_rings_coords.value.match(/,/)) {
			var coords = document.range_rings.range_rings_coords.value.split(",");
			if (coords.length) {
				document.range_rings.lat1.value = parseCoordinate(coords[0]);
				document.range_rings.lon1.value = parseCoordinate(coords[1]);
			}
		} else if (document.range_rings.range_rings_coords.value.match(/[0-9]/)) {
			alert ('The latitude and longitude need to be separated by a comma.');
		}
	}
}

function Calculate_Bearing_Form() {
	if (document.bearing.bearing_coords1.value.match(/,/)) {
		var coords1 = document.bearing.bearing_coords1.value.split(",");
		if (coords1.length) {
			document.bearing.lat1.value = parseCoordinate(coords1[0]);
			document.bearing.lon1.value = parseCoordinate(coords1[1]);
			if (document.bearing.distance.value.match(/\w/) && document.bearing.bearing.value.match(/\w/)) {
				var coords2 = Point_At_Distance_And_Bearing(document.bearing.lat1.value,document.bearing.lon1.value,document.bearing.distance.value,document.bearing.bearing.value);
				if (coords2.length) {
					document.bearing.lat2.value = parseCoordinate(coords2[0]);
					document.bearing.lon2.value = parseCoordinate(coords2[1]);
					document.bearing.bearing_coords2.value = document.bearing.lat2.value+', '+document.bearing.lon2.value;
					if (parseDistance_units == 'us') { document.bearing.units.value = 'us'; }
					else if (parseDistance_units == 'nautical') { document.bearing.units.value = 'nautical'; }
					if($('bearing_convert')) { $('bearing_convert').style.display = 'block'; }
					return true;
				}
			}
		}
	}
	return false;
}


function ChangeSubmitButton(form_id,format_id,submit_id) {
	if ($(form_id) && $(format_id) && $(format_id).value && $(submit_id)) {
		var f = $(form_id);
		if ($(format_id).value.match(/profile/)) {
			if (f.add_elevation) { f.add_elevation.value = 'auto'; }
			if (f.trk_colorize) { f.trk_colorize.value = 'altitude'; }
			if (f.gc_segments.value && parseInt(f.gc_segments.value) == 2000) { f.gc_segments.value = ''; } // why?
			f.action = '/profile?output_calculators';
			$(submit_id).value = 'Draw profile';
		} else if ($(format_id).value.match(/^(text|gpx)/)) {
			if (f.add_elevation) { f.add_elevation.value = ''; }
			if (f.trk_colorize) { f.trk_colorize.value = ''; }
			f.action = '/convert?output_calculators';
			if ($(format_id).value.match(/gpx/i)) {
				f.convert_format = 'gpx';
				$(submit_id).value = 'Create GPX file';
			} else {
				f.convert_format = 'text';
				$(submit_id).value = 'Show coordinates';
			}
		} else {
			if (f.add_elevation) { f.add_elevation.value = ''; }
			if (f.trk_colorize) { f.trk_colorize.value = ''; }
			f.action = '/map?output_calculators';
			$(submit_id).value = 'Draw map';
		}
	}
}



function uri_escape(text) {
	text = escape(text);
	text = text.replace(/\//g,"%2F");
	text = text.replace(/\?/g,"%3F");
	text = text.replace(/=/g,"%3D");
	text = text.replace(/&/g,"%26");
	text = text.replace(/@/g,"%40");
	return (text);
}




// JSONscriptRequest
//
// Author: Jason Levitt
// Date: December 7th, 2005
//
// Constructor -- pass a REST request URL to the constructor
//

function JSONscriptRequest(fullUrl) {
	// REST request path
	this.fullUrl = fullUrl; 
	// Keep IE from caching requests
	this.noCache = '&gv_nocache=' + (new Date()).getTime();
	// Get the DOM location to put the script tag
	this.headLoc = document.getElementsByTagName("head").item(0);
	// Generate a unique script tag id
	this.scriptId = 'YJscriptId' + JSONscriptRequest.scriptCounter++;
}

// Static script ID counter
JSONscriptRequest.scriptCounter = 1;

// buildScriptTag method
//
JSONscriptRequest.prototype.buildScriptTag = function () {

	// Create the script tag
	this.scriptObj = document.createElement("script");
	
	// Add script object attributes
	this.scriptObj.setAttribute("type", "text/javascript");
	this.scriptObj.setAttribute("src", (this.fullUrl.match(/&callback=/)) ? this.fullUrl.replace(/&callback=/,this.noCache+'&callback=') : this.fullUrl+this.noCache); // Google requires 'callback' to be the LAST parameter
	this.scriptObj.setAttribute("id", this.scriptId);
}
 
// removeScriptTag method
// 
JSONscriptRequest.prototype.removeScriptTag = function () {
	// Destroy the script tag
	if (this.scriptObj) { this.headLoc.removeChild(this.scriptObj); }
}

// addScriptTag method
//
JSONscriptRequest.prototype.addScriptTag = function () {
	// Create the script tag
	this.headLoc.appendChild(this.scriptObj);
}


function check_for_coordinates(address) {
	var coords = [];
	var coordinate_pattern = new RegExp('^ *([NSEW])? *((?:[0-9\.\-]+[^0-9A-Z\.\-]*)+) *([NSEW])? *, *([EWNS])? *((?:[0-9\.\-]+[^0-9A-Z\.\-]*)+) *([EWNS])? *,? *$','i');
	if (address.match(coordinate_pattern)) { // the query looks like a pair of numeric coordinates
		var coordinate_match = coordinate_pattern.exec(address.toUpperCase());
		if (coordinate_match && coordinate_match[2] != null && coordinate_match[5] != null) {
			if ((coordinate_match[1] && coordinate_match[1].match(/[EW]/i)) || (coordinate_match[3] && coordinate_match[3].match(/[EW]/i)) || (coordinate_match[4] && coordinate_match[4].match(/[NS]/i)) || (coordinate_match[6] && coordinate_match[6].match(/[NS]/i))) {
				coords[1] = parseCoordinate((coordinate_match[1]||'')+coordinate_match[2]+(coordinate_match[3]||''));
				coords[0] = parseCoordinate((coordinate_match[4]||'')+coordinate_match[5]+(coordinate_match[6]||''));
			} else {
				coords[0] = parseCoordinate((coordinate_match[1]||'')+coordinate_match[2]+(coordinate_match[3]||''));
				coords[1] = parseCoordinate((coordinate_match[4]||'')+coordinate_match[5]+(coordinate_match[6]||''));
			}
		}
	}
	if (coords[0] || coords[1]) {
		return coords;
	} else {
		return null;
	}
}



function http_post(path, params, method) {
    method = method || "post"; // Set method to post by default if not specified.
    var form = document.createElement("form");
    form.setAttribute("method", method);
    form.setAttribute("action", path);
    for(var key in params) {
        if(params.hasOwnProperty(key)) {
            var hiddenField = document.createElement("input");
            hiddenField.setAttribute("type", "hidden");
            hiddenField.setAttribute("name", key);
            hiddenField.setAttribute("value", params[key]);
            form.appendChild(hiddenField);
        }
    }
    document.body.appendChild(form);
    form.submit();
}