var whosonfirst = whosonfirst || {};
whosonfirst.spatial = whosonfirst.spatial || {};

whosonfirst.spatial.placetypes = (function(){
    
    var self = {

	init: function(){
	    
	    return new Promise((resolve, reject) => {

		var placetypes_el = document.getElementById("placetypes");

		if (! placetypes_el){
		    reject("Missing placetypes element");
		    return;
		}
		
		whosonfirst.spatial.api.placetypes({}).then((rsp) => {

		    var mk_checkbox = function(id, name){
			
			var div = document.createElement("div");
			div.setAttribute("class", "form-check form-check-inline");
			
			var input = document.createElement("input");
			input.setAttribute("type", "checkbox");
			input.setAttribute("class", "form-check-input spatial-filter spatial-filter-placetype");
			input.setAttribute("id", "placetype-" + id);
			input.setAttribute("value", name);
			
			var label = document.createElement("label");
			label.setAttribute("class", "form-check-label");
			label.setAttribute("for", "placetype-" + name);
			label.appendChild(document.createTextNode(name));
			
			div.appendChild(input);
			div.appendChild(label);
			
			return div;
		    };
		    
		    
		    var count = rsp.length;
		    
		    for (var i=0; i < count; i++){
			var pt = rsp[i];
			var cb = mk_checkbox(pt.id, pt.name);
			placetypes_el.appendChild(cb);
		    }

		    resolve();
		    
		}).catch((err) => {
		    reject(err);
		});
		
	    });
	}
	
    };

    return self;
    
})();
