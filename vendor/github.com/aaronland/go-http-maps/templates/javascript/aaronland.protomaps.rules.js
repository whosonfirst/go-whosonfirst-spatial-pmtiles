{{ define "rules" -}}
var aaronland = aaronland || {};
aaronland.protomaps = aaronland.protomaps || {};

aaronland.protomaps.rules = (function(){

    var self = {
	'paintRules': function(){
	    return {{ .PaintRules }};
	},

	'labelRules': function(){
	    return {{ .LabelRules }};
	},
    };

    return self
})();   
{{ end }}
