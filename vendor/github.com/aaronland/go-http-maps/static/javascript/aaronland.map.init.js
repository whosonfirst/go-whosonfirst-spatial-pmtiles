window.addEventListener("load", function load(event){

    // stringify featurecollection (fc) and write it
    // to a <pre> element.
    var dump_geojson = function(fc) {

	// START OF wof-specific stuff
	var count = fc.features.length;
	var f;
	
	if (count == 1){
	    f = fc.features[0];
	} else {

	    var geoms = [];

	    for (var i=0; i < count; i++){
		var f = fc.features[i];
		geoms.push(f.geometry);
	    }
	    
	    f = {
		"Type": "Feature",
		"geometry:": {
		    "Type": "MultiGeometry", geometries: geoms
		},
		"properties": {}
	    }
	}
	
	f.properties["wof:id"] = -1;
	f.properties["wof:name"] = "test";
	f.properties["wof:placetype"] = "custom";	    
	
	var str_f = JSON.stringify(f);
	
	// END OF wof-specific stuff
	
	var enc_fc = JSON.stringify(fc, "", " ");

	var pre = document.getElementById("geojson");
	pre.innerHTML = "";
	pre.appendChild(document.createTextNode(enc_fc));
    };
    
    var map_el = document.getElementById("map");
    var map = aaronland.maps.getMap(map_el);

    var hash = new L.Hash(map);
    var hash_str = location.hash;

    var init_lat = map_el.getAttribute("data-initial-latitude");
    var init_lon = map_el.getAttribute("data-initial-longitude");
    var init_zoom = map_el.getAttribute("data-initial-zoom");
    
    if (hash_str){

	var parsed = aaronland.maps.parseHash(hash_str);

	if (parsed){
	    init_lat = parsed['latitude'];
	    init_lon = parsed['longitude'];
	    init_zoom = parsed['zoom'];
	}
    }

    map.setView([init_lat, init_lon], init_zoom);

    if (map.pm){
	
	 var on_update = function(){
	     var feature_group = map.pm.getGeomanLayers(true);
	     var feature_collection = feature_group.toGeoJSON();
	     dump_geojson(feature_collection);	     
	 };
	 
	 map.pm.addControls({  
	     position: 'topleft',  
	 });
	 
	 map.on("pm:drawend", function(e){
	     console.log("draw end");
	     on_update();
	 });
	 
	 map.on('pm:remove', function (e) {
	     console.log("remove");	     
	     on_update();	     
	 });

	 // This does not appear to capture drag or edit-vertex events
	 // Not sure what's up with that...
	 
	 map.on('pm:globaleditmodetoggled', (e) => {
	     console.log("remove");	     
	     on_update();	     
	 });
    }
    
    
});
