window.addEventListener("load", function load(event){
    
    var pip_wrapper = document.getElementById("point-in-polygon");

    if (! pip_wrapper){
	console.log("Missing 'point-in-polygon' element.");
	return;
    }
    
    var init_lat = pip_wrapper.getAttribute("data-initial-latitude");

    if (! init_lat){
	console.log("Missing initial latitude");
	return;
    }
    
    var init_lon = pip_wrapper.getAttribute("data-initial-longitude");

    if (! init_lon){
	console.log("Missing initial longitude");
	return;
    }
    
    var init_zoom = pip_wrapper.getAttribute("data-initial-zoom");    

    if (! init_zoom){
	console.log("Missing initial zoom");
	return;
    }

    var max_bounds = pip_wrapper.getAttribute("data-max-bounds");
    
    var map_el = document.getElementById("map");

    if (! map_el){
	console.log("Missing map element");	
	return;
    }

    /* github.com/aaronland/go-http-maps */
    var map = aaronland.maps.getMap(map_el);
    
    if (! map){
	console.log("Unable to instantiate map");
	return;
    }
    
    var layers = L.layerGroup();
    layers.addTo(map);
    
    var spinner = new L.Control.Spinner();
    // map.addControl(spinner);

    
    var update_map = function(e){

	var pos = map.getCenter();	

	var args = {
	    'latitude': pos['lat'],
	    'longitude': pos['lng'],
	};
	
	var properties = [];

	var extra_properties = document.getElementById("extras");

	if (extra_properties){

	    var extras = extra_properties.value;

	    if (extras){
		properties = extras.split(",");
		args['properties'] = properties;
	    }
	}
	
	var existential_filters = document.getElementsByClassName("point-in-polygon-filter-existential");
	var count_existential = existential_filters.length;

	for (var i=0; i < count_existential; i++){

	    var el = existential_filters[i];

	    if (! el.checked){
		continue;
	    }
	    
	    var fl = el.value;
	    args[fl] = [ 1 ];
	}

	var placetypes = [];
	
	var placetype_filters = document.getElementsByClassName("point-in-polygon-filter-placetype");	
	var count_placetypes = placetype_filters.length;

	for (var i=0; i < count_placetypes; i++){

	    var el = placetype_filters[i];

	    if (! el.checked){
		continue;
	    }

	    var pt = el.value;
	    placetypes.push(pt);
	}

	if (placetypes.length > 0){
	    args['placetypes'] = placetypes;
	}

	var edtf_filters = document.getElementsByClassName("point-in-polygon-filter-edtf");
	var count_edtf = edtf_filters.length;

	for (var i=0; i < count_edtf; i++){

	    var el = edtf_filters[i];

	    var id = el.getAttribute("id");

	    if (! id.match("^(inception|cessation)$")){
		continue
	    }

	    var value = el.value;

	    if (value == ""){
		continue;
	    }
	    
	    // TO DO: VALIDATE EDTF HERE WITH WASM
	    // https://millsfield.sfomuseum.org/blog/2021/01/14/edtf/

	    var key = id + "_date";
	    args[key] = value;
	};

	var show_feature = function(id){

	    var data_root = document.body.getAttribute("data-root");

	    if (!data_root.endsWith("/")){
		data_root = data_root + "/";
	    }
	    
	    var url = data_root + id;

	    var on_success = function(data){

		var l = L.geoJSON(data, {
		    style: function(feature){
			return whosonfirst.spatial.pip.named_style("match");
		    },
		});
		
		layers.addLayer(l);
		l.bringToFront();
	    };

	    var on_fail= function(err){
		console.log("SAD", id, err);
	    }
	    
	    whosonfirst.net.fetch(url, on_success, on_fail);
	};
	
	var on_success = function(rsp){

	    map.removeControl(spinner);
	    
	    var places = rsp["places"];
	    var count = places.length;

	    var matches = document.getElementById("pip-matches");
	    matches.innerHTML = "";
	    
	    if (! count){
		return;
	    }
	    
	    for (var i=0; i < count; i++){
		var pl = places[i];
		show_feature(pl["wof:id"]);
	    }
	    
	    var table_props = whosonfirst.spatial.pip.default_properties();

	    // START OF something something something
	    
	    var extras_el = document.getElementById("extras");

	    if (extras_el){
		
		var str_extras = extras_el.value;
		var extras = null;
		
		if (str_extras){
		    extras = str_extras.split(",");  		    
		}

		if (extras){

		    var first = places[0];
		    
		    var count_extras = extras.length;		    
		    var extra_props = [];
		    
		    for (var i=0; i < count_extras; i++){

			var ex = extras[i];
			
			if ((ex.endsWith(":")) || (ex.endsWith(":*"))){
			    
			    var prefix = ex.replace("*", "");
			    
			    for (k in first){
				if (k.startsWith(prefix)){
				    extra_props.push(k);
				}
			    }
			    
			} else {

			    if (first[ex]) {
				extra_props.push(ex);
			    }
			}
		    }

		    for (idx in extra_props){
			var ex = extra_props[idx];
			table_props[ex] = "";
		    }
		}

	    }

	    // END OF something something something
	    
	    var table = whosonfirst.spatial.pip.render_properties_table(places, table_props);
	    matches.appendChild(table);
	    
	};

	var on_error = function(err){

	    var matches = document.getElementById("pip-matches");
	    matches.innerHTML = "";
	    
	    map.removeControl(spinner);	    
	    console.log("SAD", err);
	}

	args["sort"] = [
	    "placetype://",
	    "name://",
	    "inception://",
	];
	
	whosonfirst.spatial.api.point_in_polygon(args, on_success, on_error);

	map.addControl(spinner);	
	layers.clearLayers();	
    };
    
    map.on("moveend", update_map);

    var filters = document.getElementsByClassName("point-in-polygon-filter");
    var count_filters = filters.length;
    
    for (var i=0; i < count_filters; i++){	    
	var el = filters[i];
	el.onchange = update_map;
    }

    var extras = document.getElementsByClassName("point-in-polygon-extra");
    var count_extras = extras.length;
    
    for (var i=0; i < count_extras; i++){	    
	var el = extras[i];
	el.onchange = update_map;
    }
    
    var hash_str = location.hash;

    if (hash_str){

	/* github.com/aaronland/go-http-maps */	
	var parsed = aaronland.maps.parseHash(hash_str);	

	if (parsed){
	    init_lat = parsed['latitude'];
	    init_lon = parsed['longitude'];
	    init_zoom = parsed['zoom'];
	}
    }
    
    map.setView([init_lat, init_lon], init_zoom);    

    if (max_bounds) {

	var bounds = max_bounds.split(",");

	var miny;
	var minx;
	var maxy;
	var maxx;
	
	if (bounds.length == 4){
	    minx = parseFloat(bounds[0]);
	    miny = parseFloat(bounds[1]);	    
	    maxx = parseFloat(bounds[2]);
	    maxy = parseFloat(bounds[3]);	    
	}

	if ((miny) && (minx) && (maxy) && (maxx)){

	    var max_bounds = [
		[ miny, minx ],
		[ maxy, maxx ]
	    ];

	    // console.log("BOUNDS", bounds, max_bounds);
	    map.setMaxBounds(max_bounds);
	}
    }
    
    slippymap.crosshairs.init(map);    
});
