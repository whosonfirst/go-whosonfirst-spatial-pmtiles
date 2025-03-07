var whosonfirst = whosonfirst || {};
whosonfirst.spatial = whosonfirst.spatial || {};

whosonfirst.spatial.api = (function(){

    var self = {

	'point_in_polygon': function(args) {

	    var rel_url = "/api/point-in-polygon";
	    return self.post(rel_url, args)
	},

	'intersects': function(args) {

	    var rel_url = "/api/intersects";
	    return self.post(rel_url, args)
	},
	
	'placetypes': function(args) {

	    var rel_url = "/api/placetypes";
	    return self.post(rel_url, args)
	},
	
	'post': function(rel_url, args) {

	    console.debug("Execute API request", rel_url, args);
	    
	    return new Promise((resolve, reject) => {
		
		var abs_url = self.abs_url(rel_url);
		
		var req = new XMLHttpRequest();
		
		req.onload = function(){
		    
		    var rsp;
		    
		    try {
			rsp = JSON.parse(this.responseText);
            	    }
		    
		    catch (e){
			console.log("ERR", abs_url, e);
			reject(e);
			return false;
		    }

		    resolve(rsp);
       		};
	    
		req.open("POST", abs_url, true);
		
		// See this? This is not great. I am still trying to figure things out. See also:
	        // https://github.com/whosonfirst/go-whosonfirst-spatial-pip/blob/main/api/http.go
		// (20210325/thisisaaronland)
		
		if (args["properties"]){
		    str_props = args["properties"].join(",");
		    req.setRequestHeader("X-Properties", str_props);
		    delete(args["properties"]);
		}
	    
		var enc_args = JSON.stringify(args);
		req.send(enc_args);	    
	    });
	},
	
	'abs_url': function(rel_url) {
	    return location.protocol + "//" + location.host + rel_url;
	},

	'query_string': function(args){

	    var pairs = [];

	    for (var k in args){

		var v = args[k];

		var enc_k = encodeURIComponent(k);
		var enc_v = encodeURIComponent(v);
		
		var pair = enc_k + "=" + enc_v;
		pairs.push(pair);
	    }

	    return pairs.join("&");
	},
    };

    return self;
    
})();
