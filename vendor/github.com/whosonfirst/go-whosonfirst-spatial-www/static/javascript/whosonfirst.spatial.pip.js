var whosonfirst = whosonfirst || {};
whosonfirst.spatial = whosonfirst.spatial || {};

whosonfirst.spatial.pip = (function(){

    var styles = {
	"match": {
	    "color": "#000",
	    "weight": 1,
	    "opacity": 1,
	    "fillColor": "#00308F",
	    "fillOpacity": 0.05
	}
    };
    
    var self = {

	'named_style': function(name){
	    return styles[name];
	},
	
	'default_properties': function(){

	    var props_table = {
		"wof:id":"",
		"wof:name":"",
		"wof:placetype":"",
		"edtf:inception":"",
		"edtf:cessation":"",		
	    };

	    return props_table;
	},
	
	'render_properties_table': function(features, props_table){

	    if (! props_table){
		props_table = self.default_properties();
	    }
	    
	    var count = features.length;
	    
	    var table = document.createElement("table");
	    table.setAttribute("class", "table table-striped");	   
	    
	    for (var i=0; i < count; i++){

		var props = features[i];
		
		// draw table header

		if (i % 10 == 0){

		    var tr = document.createElement("tr");
	    
		    for (var k in props_table){
			
			if (self.is_wildcard(k)){
			    
			    for (prop_k in props){
				
				if (! prop_k.startsWith(k)){
				    continue;
				}
				
				var v = prop_k;
				
				var th = document.createElement("th");
				th.appendChild(document.createTextNode(v));
				tr.appendChild(th);				
			    }
			    
			} else {
			    
			    var v = k;	// props_table[k]
			    var th = document.createElement("th");
			    th.appendChild(document.createTextNode(v));
			    tr.appendChild(th);			    
			}		
		    }
		    
		    var thead = document.createElement("thead");
		    thead.setAttribute("class", "thead-dark");
		    thead.appendChild(tr);
		    table.appendChild(thead);		    
		}
		
		var wof_id = props["wof:id"];
		
		var tr = document.createElement("tr");
		tr.setAttribute("id", "tr-" + wof_id);
		
		for (var k in props_table){

		    if (self.is_wildcard(k)){

			for (prop_k in props){

			    if (! prop_k.startsWith(k)){
				continue;
			    }

			    var v = props[prop_k];
			    var node = self.render_value(v);
			    
			    var td = document.createElement("td");
			
			    td.appendChild(node);
			    tr.appendChild(td);
			}
			
		    } else {
			
			var v = props[k];
			var node = self.render_value(v);

			var td = document.createElement("td");
			
			td.appendChild(node);
			tr.appendChild(td);
		    }
		    
		    table.appendChild(tr);
		}
		
	    }

	    var wrapper = document.createElement("div");
	    wrapper.setAttribute("class", "table-responsive");

	    wrapper.appendChild(table);
	    return wrapper;
	},

	'is_wildcard': function(str) {

	    if (str.endsWith(":")){
		return true;
	    }
	    
	    if (str.endsWith("*")){
		return true;
	    }

	    return false;
	},

	'render_value': function(v) {

	    if (typeof(v) == "object"){

		var enc_v = JSON.stringify(v, null, 2);
		var pre = document.createElement("pre");
		pre.appendChild(document.createTextNode(enc_v));

		var summary = document.createElement("summary");
		summary.appendChild(document.createTextNode("details"));
		    
		var details = document.createElement("details");
		details.appendChild(summary);
		details.appendChild(pre);
		
		return details;
	    }

	    else {
		return document.createTextNode(v);
	    }
	},
    };

    return self;
    
})();
