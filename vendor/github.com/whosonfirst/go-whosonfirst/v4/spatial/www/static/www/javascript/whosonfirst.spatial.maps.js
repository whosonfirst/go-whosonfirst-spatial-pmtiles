var whosonfirst = whosonfirst || {};
whosonfirst.spatial = whosonfirst.spatial || {};

whosonfirst.spatial.maps = (function(){
    
    var self = {

	init: function(){
	    
	    return new Promise((resolve, reject) => {

		// Null Island    
		var map = L.map('map').setView([0.0, 0.0], 1);    
		
		fetch("/map.json")
		    .then((rsp) => rsp.json())
		    .then((cfg) => {
			
			console.debug("Got map config", cfg);
			
			switch (cfg.provider) {
			    case "leaflet":
				
				var tile_url = cfg.tile_url;
				
				var tile_layer = L.tileLayer(tile_url, {
				    maxZoom: 19,
				});
				
				tile_layer.addTo(map);
				break;
				
			    case "protomaps":
				
				var tile_url = cfg.tile_url;
				
				var tile_layer = protomapsL.leafletLayer({
				    url: tile_url,
				    theme: cfg.protomaps.theme,
				})
				
				tile_layer.addTo(map);
				break;
				
			    default:
				console.error("Uknown or unsupported map provider");
				reject("Invalid map provider");				
				return;
			}
			
			if (cfg.initial_view) {
			    
			    var zm = map.getZoom();
			    
			    if (cfg.initial_zoom){
				zm = cfg.initial_zoom;
			    }
			    
			    map.setView([cfg.initial_view[1], cfg.initial_view[0]], zm);
			    
			} else if (cfg.initial_bounds){
			    
			    var bounds = [
				[ cfg.initial_bounds[1], cfg.initial_bounds[0] ],
				[ cfg.initial_bounds[3], cfg.initial_bounds[2] ],
			    ];
			    
			    map.fitBounds(bounds);
			}
			
			console.debug("Finished map setup");
			resolve(map);
	    
		    }).catch((err) => {
			console.error("Failed to derive map config", err);
			reject(err);
			return;
		    });    
		
	    });
	}
	
    };

    return self;
    
})();
