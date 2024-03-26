var aaronland = aaronland || {};

aaronland.maps = (function(){

    var maps = {};

    var self = {

	'parseHash': function(hash_str){

	    if (hash_str.indexOf('#') === 0) {
		hash_str = hash_str.substr(1);
	    }

	    var lat;
	    var lon;
	    var zoom;
	    
	    var args = hash_str.split("/");
	    
	    if (args.length != 3){
		console.log("Unrecognized hash string");
		return null;
	    }
	    
	    zoom = args[0];
	    lat = args[1];
	    lon = args[2];			
	    
	    zoom = parseInt(zoom, 10);
	    lat = parseFloat(lat);
	    lon = parseFloat(lon);		
	    
	    if (isNaN(zoom) || isNaN(lat) || isNaN(lon)) {
		console.log("Invalid zoom/lat/lon", zoom, lat, lon);
		return null;
	    }

	    var parsed = {
		'latitude': lat,
		'longitude': lon,
		'zoom': zoom,
	    };

	    return parsed;
	},

	'getMapById': function(map_id, args){

	    var map_el = document.getElementById("map");

	    if (! map_el){
		return null;
	    }

	    return self.getMap(map_el, args);
	},
	
	'getMap': function(map_el, map_args){

	    if (! map_args){
		map_args = {};
	    }
	    
	    var map_id = map_el.getAttribute("id");

	    if (! map_id){
		return;
	    }
	    
	    if (maps[map_id]){
		return maps[map_id];
	    }

	    var map = L.map(map_id, map_args);

	    var map_provider = map_el.getAttribute("data-map-provider");

	    switch (map_provider) {

		case "leaflet":

		    var tile_url = document.body.getAttribute("data-leaflet-tile-url");

		    var layer = L.tileLayer(tile_url);
		    layer.addTo(map);
		    break;
		    
		case "protomaps":
		
		    var tile_url = document.body.getAttribute("data-protomaps-tile-url");

		    var args = {
			url: tile_url,
		    };

		    var paint_rules = aaronland.protomaps.rules.paintRules();
		    var label_rules = aaronland.protomaps.rules.labelRules();
		    
		    if (paint_rules){
			args['paint_rules'] = paint_rules;
		    }
		   
		    if (label_rules){
			args['label_rules'] = label_rules;
		    }
		    
		    var layer = protomaps.leafletLayer(args)
		    layer.addTo(map);
		    break;
		    
		case "tangram":
		    
		    map.setMaxZoom(17.99);	// TO DO: make this Z20 or something...
		    
		    var tangram_opts = self.getTangramOptions();	   
		    var tangramLayer = Tangram.leafletLayer(tangram_opts);
		    
		    tangramLayer.addTo(map);
		case "null":
		    break

		default:
		    console.log("Unsupported map provider ", map_provider);
	    }

	    var attribution = self.getAttribution(map_provider);
	    map.attributionControl.addAttribution(attribution);
	    
	    maps[map_id] = map;
	    return map;
	},

	'getTangramOptions': function(){

	    var api_key = document.body.getAttribute("data-nextzen-api-key");
	    var style_url = document.body.getAttribute("data-nextzen-style-url");
	    var tile_url = document.body.getAttribute("data-nextzen-tile-url");
	    
	    /*
	    var sceneText = await fetch(new Request('https://somwehere.com/scene.zip', { headers: { 'Accept': 'application/zip' } })).then(r => r.text());
	    var sceneURL = URL.createObjectURL(new Blob([sceneText]));
	    scene.load(sceneURL, { base_path: 'https://somwehere.com/' });
	    */
	    
	    var tangram_opts = {
		scene: {
		    import: [
			style_url,
		    ],
		    sources: {
			mapzen: {
			    url: tile_url,
			    url_subdomains: ['a', 'b', 'c', 'd'],
			    url_params: {api_key: api_key},
			    tile_size: 512,
			    max_zoom: 18
			}
		    }
		}
	    };

	    return tangram_opts;
	},

	'getAttribution': function(map_provider){

	    if (map_provider == "tangramjs"){
		return '<a href="https://github.com/tangrams" target="_blank">Tangram</a> | <a href="http://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a> | <a href="https://www.nextzen.org/" target="_blank">Nextzen</a>';
	    }

	    if (map_provider == "protomaps"){
		return '<a href="https://github.com/protomaps" target="_blank">Protomaps</a> | <a href="http://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>';
	    }
	    
	    return '<a href="http://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>';
	},
    };

    return self;
    
})();
