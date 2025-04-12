var whosonfirst = whosonfirst || {};
whosonfirst.spatial = whosonfirst.spatial || {};

whosonfirst.spatial.piptile = (function(){
    
    var self = {
	
	init: function(map) {

	    var zoom_el = document.getElementById("at_zoom");
	    
	    var layers = L.layerGroup();
	    layers.addTo(map);
	    
	    var spinner = new L.Control.Spinner();
	    
	    var update_map = function(e){
		
		var pos = map.getCenter();	

		console.debug("Map center", pos);

		var zm = parseInt(zoom_el.value);

		console.debug("derive tile", pos, zm);
		
		var tile = self.tileAt(pos, zm)
		console.debug("Fetch for tile", tile);
		
		var args = {
		    tile: tile,
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
		
		var existential_filters = document.getElementsByClassName("spatial-filter-existential");
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
		
		var placetype_filters = document.getElementsByClassName("spatial-filter-placetype");	
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
		
		var edtf_filters = document.getElementsByClassName("spatial-filter-edtf");
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

		    var url = "/data/" + id;
		    
		    var on_success = function(data){
			
			var l = L.geoJSON(data, {
			    style: function(feature){
				return whosonfirst.spatial.results.named_style("match");
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

		    var l = L.geoJSON(rsp, {
			    style: function(feature){
				return whosonfirst.spatial.results.named_style("match");
			    },
		    });

		    layers.addLayer(l);
		    l.bringToFront();
		    
		    var features = rsp["features"];
		    var count = features.length;
		    
		    var matches = document.getElementById("pip-matches");
		    matches.innerHTML = "";
		    
		    if (! count){
			return;
		    }

		    var places = [];

		    for (var i=0; i < count; i++){
			places[i] = features[i].properties;
		    }
		    
		    /*
		    for (var i=0; i < count; i++){
			var pl = places[i];
			show_feature(pl["wof:id"]);
		    }
		     */
		    
		    var table_props = whosonfirst.spatial.results.default_properties();
		    
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
		    
		    var table = whosonfirst.spatial.results.render_properties_table(places, table_props);
		    matches.appendChild(table);
		    
		};
		
		var on_error = function(err){
		    
		    var matches = document.getElementById("pip-matches");
		    matches.innerHTML = "";
		    
		    map.removeControl(spinner);	    
		    console.error("Point in polygon request failed", err);
		}
		
		args["sort"] = [
		    "placetype://",
		    "name://",
		    "inception://",
		];
		
		whosonfirst.spatial.api.point_in_polygon_with_tile(args).then((rsp) => {
		    on_success(rsp);
		}).catch((err) => {
		    on_error(err);
		});
		
		map.addControl(spinner);	
		layers.clearLayers();	
	    };
	    
	    map.on("moveend", update_map);

	    zoom_el.onchange = update_map;
	    
	    var filters = document.getElementsByClassName("spatial-filter");
	    var count_filters = filters.length;
	    
	    for (var i=0; i < count_filters; i++){	    
		var el = filters[i];
		el.onchange = update_map;
	    }
	    
	    var extras = document.getElementsByClassName("spatial-extra");
	    var count_extras = extras.length;
	    
	    for (var i=0; i < count_extras; i++){	    
		var el = extras[i];
		el.onchange = update_map;
	    }
	    	    
	    slippymap.crosshairs.init(map);

	    whosonfirst.spatial.placetypes.init().catch((err) => {
		console.error("Failed to initialize placetypes", err);
	    });
	    
	},

	tileAt: function(pos, zm) {
	    const coords = self.fraction(pos, zm);
	    return { x: Math.floor(coords.x), y: Math.floor(coords.y), zoom: zm }
	},

	// https://github.com/paulmach/orb/blob/v0.11.1/maptile/tile.go#L143
	
	fraction: function(pos, zm) {

	    var x;
	    var y;

	    // Oh Javascript...
	    const lon = Number(pos.lng.toPrecision(7));
	    const lat = Number(pos.lat.toPrecision(7));
	    const pi = Number(Math.PI.toPrecision(7));
	    
	    const factor = 1 << zm;
	    console.log("factor", factor)
	    
	    const maxtiles = parseFloat(factor);
	    const tmp_lon = lon / 360.0 + 0.5;
	    
	    x = tmp_lon * maxtiles;
	    
	    if (lat < -85.0511) {
		y = maxtiles - 1;
	    } else if (lat > 85.0511) {
		y = 0;
	    } else {

		const siny = Math.sin(lat * pi / 180.0);
		const tmp_lat = 0.5 + 0.5 * Math.log((1.0 + siny)/(1.0 - siny))/(-2 * pi)
		
		y = tmp_lat * maxtiles;
	    }

	    return { 'x': x, 'y': y };
	},
    };

    return self;
    
})();
