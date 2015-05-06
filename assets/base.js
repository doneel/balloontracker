var paths = [];
$(reqData);

function reqData(){
    console.log('asking');
    var pathsReq = $.get('/paths/', loadData);
}

var loadData = function(data) {
    paths = data;
    console.log(paths);    
    var alts = {};
    $.each(paths, function (index, value) {
        alts[value["Ceiling"]] = 1
    });
    console.log(alts)
    $.each(alts, function(key, value) {
            $('#altitude-selector').append($('<option/>', { 
                        value: key,
                        text : key
        }));
    });
    showHeatMap(paths);
}


function showHeatMap(paths) {
    heatmapData = [];
    $.each(paths, function(ind, path){
        heatmapData.push(new google.maps.LatLng(path.EndLat, path.EndLon));
    });
    console.log(heatmapData)
    var heatmap = new google.maps.visualization.HeatmapLayer({
          data: heatmapData
    });
    heatmap.setMap(map);
}


function initialize() {
            var mapOptions = {
              center: new google.maps.LatLng(45.220, -111.761),
              zoom: 8
            };
            map = new google.maps.Map(document.getElementById("map-canvas"), mapOptions);
}
google.maps.event.addDomListener(window, 'load', initialize);
            /*
           displayedLines = [];
           for(var i = pathsList.length - 1; i >= 0; i--){
                flight = pathsList[i];
                fullObj = buildFlightRepresentation(flight, new_map);
                displayedLines[i] = fullObj

                $(".flightListBox").append(fullObj.box);

           }


/*
            var longs = [-111.761, -111.601, -110.785, -110.074]
            var latits = [45.220, 45.268, 44.873, 44.082]

            var ll_arr = [];
            for(var i = 0; i < latits.length; i++){
              console.log(latits[i], ' ', longs[i])
              ll_arr[i] = new google.maps.LatLng(latits[i], longs[i]);
            }
            console.log(ll_arr);

            var Longpath = new google.maps.MVCArray(ll_arr);
            console.log(Longpath);



            var line = new google.maps.Polyline(lineOptions);
            */
//          }


        /*
           $(document).ready(function(){
                  $("#linesBox").on('change', function(evt){
                        showPaths(this.checked);
                 });

                 $("#markerBox").change(function(){
                        showMarkers(this.checked);
                 });

                 $("#opacityBox").change(function(){
                        scaleOpacities(this.checked);
                 });

            });

           function showMarkers(should){
                var opts = {
                    visible: should
                }
                for(var i = 0; i < displayedLines.length; i++) {
                    displayedLines[i].marker.setOptions(opts);
                }
           }

           function showPaths(should){
              var opts = {
                visible: should
              }
                for(var i = 0; i < displayedLines.length; i++) {
                    displayedLines[i].path.setOptions(opts);
                }
           }

           function scaleFunction(x){
                return Math.pow(Math.E, x * Math.log(.25) * 2);
           }

           function scaleOpacities(should){
                var opts = {
                  opacity: 1
                }

                /*var shift = 1 - (displayedLines.length - 1)/displayedLines.length
                var maxVal = scaleFunction((displayedLines.length - 1)/displayedLines.length)
                console.log('Max Value ', maxVal)
                for(var i = 0; i < displayedLines.length ; i++) {
                    closenessToEnd = (displayedLines.length - i) / displayedLines.length + shift;
                    opacity = scaleFunction(closenessToEnd)/maxVal;
                    console.log(opacity);
                    displayedLines[i].path.setOptions({strokeOpacity: !should || opacity}); //janky?
                    displayedLines[i].marker.setOptions(opts);
                }
                */
/*
                var newest = displayedLines[displayedLines.length - 1].date;
                for(var i = 0; i < displayedLines.length; i++){
                      var timeGap = newest - displayedLines[i].date;
                      var dayGap = timeGap/(1000 * 60 * 60 * 24);
                      var decayFactor = Math.pow(.5, dayGap/30);
                      var opacity = 1 * decayFactor;
                      console.log(opacity);
                      displayedLines[i].path.setOptions({strokeOpacity: !should || opacity}); //janky?
                      var curColor = displayedLines[i].marker.icon.strokeColor;
                      displayedLines[i].marker.setOptions({icon: getIcon(curColor, !should || opacity)});
                }
           }

           function getIcon(color, opacity){
               options = {
                          path: google.maps.SymbolPath.CIRCLE,
                          scale: 5,
                          strokeColor: color,
                          strokeOpacity: .5
                      };
                      return options;
           }

           function buildFlightRepresentation(flight, map){
                ll_path = [];
                for(var reading = 0; reading < flight.Readings.length; reading++){
                      ll_path[reading] = new google.maps.LatLng(flight.Readings[reading].Latitude, flight.Readings[reading].Longitude)
                }
                var displayPath = new google.maps.MVCArray(ll_path)
                var lineOptions = {
                  path : displayPath,
                  map : map,
                  editable: false,
                  draggable: false,
                  strokeColor: '#DD5522',
                  strokeOpacity: 1

                };
                var path = new google.maps.Polyline(lineOptions);


                var marker = new google.maps.Marker({position: new google.maps.LatLng(flight.FinalLat, flight.FinalLong), icon: {
                          path: google.maps.SymbolPath.CIRCLE,
                          scale: 5,
                          strokeColor: '#CC1144',
                          strokeOpacity: .5
                      } , map: map});

                google.maps.event.addListener(path, 'mouseover', function(event){
                     if(this.hilighted === true){
                          console.log('dont fire!')
                          return;
                     }
                     this.hilighted = true;
                     this.previousOpacity = this.strokeOpacity;
                     this.previousColor = this.strokeColor;
                     this.setOptions({strokeOpacity: 1, editable: false, strokeColor: '#1C8A00'});
                      console.log(marker.icon)
                      marker.prevColor = marker.icon.strokeColor;
                      marker.prevOpacity = marker.icon.strokeOpacity;
                      marker.setOptions({
                        icon: getIcon('#1C8A00', 1)
                      });
                });

                google.maps.event.addListener(path, 'mouseout', function(event){
                      this.hilighted = false;
                      opacity = this.previousOpacity;
                      color = this.previousColor;
                      console.log(this.previousOpacity)
                     this.setOptions({strokeOpacity: opacity , editable: false, strokeColor: color});
                     marker.setOptions({
                          icon: getIcon(marker.prevColor, marker.prevOpacity)
                     })
                });

                var containerDiv = document.createElement('div');
                containerDiv.className = "flightBox";

                $(containerDiv).on('click', {lat: flight.FinalLat, lon: flight.FinalLong, map: map}, function(evt){
                          evt.data.map.setOptions({center: new google.maps.LatLng(evt.data.lat, evt.data.lon), zoom: 10});
                });


                var dateP = document.createElement('p');
                dateP.className = "flightTimeBox";

                var date = new Date(flight.Timestamp);
                dateP.innerHTML = (date.getMonth()+1) + "/" + date.getDate() + "/" + date.getFullYear() + " " + date.getHours() + ":00"

                containerDiv.appendChild(dateP);


                var context = this;

                $(containerDiv).on('mouseover', {path:path}, function(evt){
                     google.maps.event.trigger(evt.data.path, 'mouseover');
                });
                $(containerDiv).on('mouseout', {path:path}, function(evt){
                     google.maps.event.trigger(evt.data.path, 'mouseout');
                });

                google.maps.event.addListener(marker, 'mouseover', function(event){
                       google.maps.event.trigger(path, 'mouseover') ;
                });
                google.maps.event.addListener(marker, 'mouseout', function(event){
                       google.maps.event.trigger(path, 'mouseout') ;
                });


                flightRepresentation = {path: path, marker: marker, box: containerDiv, date: date}
                return flightRepresentation;

           }
*/
