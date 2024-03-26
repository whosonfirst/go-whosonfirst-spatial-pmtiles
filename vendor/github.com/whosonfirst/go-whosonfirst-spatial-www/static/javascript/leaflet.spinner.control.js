L.Control.Spinner = L.Control.extend({

    options: {
        position: 'topright',
    },

    onAdd: function (map) {
	
        var container = L.DomUtil.create('div', 'leaflet-control-image leaflet-bar leaflet-control');

	// <div id="spinner" style="display:none;"><div class="spinner"></div></div>

	var spinner = L.DomUtil.create('div', '', container);
	spinner.setAttribute('id', 'spinner');

	var canvas = L.DomUtil.create('div', 'spinner', spinner);
	
	this.spinner = spinner;
	
        // L.DomEvent.on(this.galleries_link, 'click', this._galleries, this);

	// This is important - without it clicking on a control in
	// rapid succcession will be interpretted as a double-click
	// causing the map to zoom (20210311/thisisaaronland)
	
	L.DomEvent.disableClickPropagation(container);
	
	return container;
    },
    
    onRemove: function(map) {
	// 
    },
    
});    
