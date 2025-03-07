window.addEventListener("load", function load(event){

    whosonfirst.spatial.maps.init().then((map) => {
	whosonfirst.spatial.intersects.init(map);	
    }).catch((err) => {
	console.error("Failed to initialize map", err)
    });

});
