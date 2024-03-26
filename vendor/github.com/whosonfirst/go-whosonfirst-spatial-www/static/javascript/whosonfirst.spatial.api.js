var whosonfirst = whosonfirst || {};
whosonfirst.spatial = whosonfirst.spatial || {};

/*

This is really a whosonfirst-spatial-pip API right now. It is too
soon to say whether this reflect a common approach for all API-related
stuff (20210322/thisisaaronland)

*/

whosonfirst.spatial.api = (function(){

    var self = {

	'point_in_polygon': function(args, on_success, on_error) {

	    var rel_url = "/point-in-polygon";
	    return self.post(rel_url, args, on_success, on_error);
	},

	'point_in_polygon_candidates': function(args, on_success, on_error) {

	    return self.post(rel_url, args, on_success, on_error);
	},

	'post': function(rel_url, args, on_success, on_error) {

	    var abs_url = self.abs_url(rel_url);
	    
	    var req = new XMLHttpRequest();
					    
	    req.onload = function(){
		
		var rsp;
		
		try {
		    rsp = JSON.parse(this.responseText);
            	}
		
		catch (e){
		    console.log("ERR", abs_url, e);
		    on_error(e);
		    return false;
		}
		
		on_success(rsp);
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
	},
	
	'abs_url': function(rel_url) {
	    var api_root = document.body.getAttribute("data-api-root");
	    return location.protocol + "//" + location.host + api_root + rel_url;
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
