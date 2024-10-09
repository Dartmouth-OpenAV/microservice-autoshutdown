package main

import (
    "github.com/labstack/echo/v4"
    "net/http"
    "time"
    "fmt"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "encoding/json"
)

// globals
var lastInteraction = make( map[string]string )
var lastInteractionMutex = &sync.Mutex{}
var deviceErrors = make( map[string]map[string]string )
var deviceErrorsMutex = &sync.Mutex{}


func registerMicroserviceFunctions( router *echo.Echo ) {
	router.GET( "/", index )
	router.GET( "/:address/occupancy_detected", getOccupancyDetected )
	router.PUT( "/:address/occupancy_detected", updateOccupancyDetected )
	router.GET( "/:address/time_avoidance", timeAvoidance )
	router.GET( "/:address/errors", getErrors )
}


func index( context echo.Context ) error {
	return context.String( http.StatusOK, "OpenAV Autoshutdown MicroService" )
}


func timeAvoidance( context echo.Context ) error {
	var key = getKey( context )
	log( key )

	from := context.QueryParam( "from" )
	if from=="" {
		from = "0730"
	}
	intFrom, err := strconv.Atoi( from )
	if err != nil {
		addToErrors( key, "couldn't convert from to int with value: " + from + ", and error: " + err.Error() )
		intFrom = 730 ;
	}

	to := context.QueryParam( "to" )
	if to=="" {
		to = "2000"
	}
	intTo, err := strconv.Atoi( to )
	if err != nil {
		addToErrors( key, "couldn't convert to to int with value: " + to + ", and error: " + err.Error() )
		intTo=2000
	}

	if( intFrom<0 ||
		intFrom>2359 ) {
		intFrom = 730 ;
		addToErrors( key, "invalid value for from: " + from + ", must be between 0 and 2359" )
	}
	if( intTo<0 ||
		intTo>2359 ) {
		intTo = 2000 ;
		addToErrors( key, "invalid value for to: " + to + ", must be between 0 and 2359" )
	}
	if( intTo<intFrom ) {
		addToErrors( key, "to is smaller than from with values to: " + to + ", and from: " + from )
		intFrom = 730 ;
		intTo = 2000 ;
	}

	// timeStamp := getTimeStamp() ;
	splitTimeStamp := strings.Split( getTimeStamp(), " " ) ;
	splitTimeStamp = strings.Split( splitTimeStamp[1], ":" ) ;
	h, err := strconv.Atoi( splitTimeStamp[0] ) ;
	m, err := strconv.Atoi( splitTimeStamp[1] ) ;
	militaryTime := h*100+m ;

	if( intFrom<=militaryTime &&
		militaryTime<=intTo ) {
		log( "current time of: " + strconv.Itoa(militaryTime) + " is between " + strconv.Itoa(intFrom) + " and " + strconv.Itoa(intTo) )
		return context.JSON( http.StatusOK, true )
	} else {
		log( "current time of: " + strconv.Itoa(militaryTime) + " is not between " + strconv.Itoa(intFrom) + " and " + strconv.Itoa(intTo) )
		return context.JSON( http.StatusOK, false )
	}

	addToErrors( key, "we shouldn't reach this point #eufghkrjgfhn" ) ;
	return context.JSON( http.StatusInternalServerError, "we shouldn't reach this point #eufghkrjgfhn" )
}



func getOccupancyDetected( context echo.Context ) error {
	var key = getKey( context )
	log( key )
	
	if _, ok := lastInteraction[key]; !ok {
		lastInteractionMutex.Lock()
		// lastInteraction[key] = getOriginTimeStamp()
		lastInteraction[key] = getTimeStamp() // safer to assume that we had an interaction when the system initialized
		lastInteractionMutex.Unlock()
	}

	lastXMinutes := context.QueryParam( "last_x_minutes" )
	if lastXMinutes=="" {
		lastXMinutes = "180"
	}
	intLastXMinutes, err := strconv.Atoi( lastXMinutes )
	if err != nil {
		addToErrors( key, "couldn't convert lastXMinutes to int with value: " + lastXMinutes + ", and error: " + err.Error() )
		intLastXMinutes = 180
	}
	log( "checking if last interaction timestamp is within " + strconv.Itoa(intLastXMinutes) + " minutes" )
	
	lastInteractionForDevice := lastInteraction[key]
	log( "last interaction timestamp: " + lastInteractionForDevice ) ;
	parsedLastInteractionForDevice, err := time.Parse( "2006-01-02 15:04:05.000000", lastInteractionForDevice )
	_ = parsedLastInteractionForDevice // appease
	if err != nil {
		addToErrors( key, "couldn't parse LastInteractionForDevice to time with value: " + lastInteractionForDevice + ", and error: " + err.Error() )
		return context.JSON( http.StatusInternalServerError, "couldn't parse LastInteractionForDevice to time with value: " + lastInteractionForDevice + ", and error: " + err.Error() )
	}

	loc, _ := time.LoadLocation( "America/New_York" )
	now, err := time.Parse( "2006-01-02 15:04:05.000000", time.Now().In(loc).Format("2006-01-02 15:04:05.000000") ) ;
	timeDifference := now.Sub( parsedLastInteractionForDevice )
	intTimeDifference := int( timeDifference.Minutes() ) ;
	log( "time difference: " + strconv.Itoa(intTimeDifference) ) 
	
	// loc, _ := time.LoadLocation( "America/New_York" )
	// timeStampXMinutesAgo := time.Now().In( loc ).Add( (-1)*time.Minute*time.Duration(intLastXMinutes) )
	// log( "timeStampXMinutesAgo: " + timeStampXMinutesAgo.Format("2006-01-02 15:04:05.000000") )
 
	if intTimeDifference<intLastXMinutes {
		return context.JSON( http.StatusOK, true )
	} else {
		return context.JSON( http.StatusOK, false )
	}
}


func updateOccupancyDetected( context echo.Context ) error {
	var key = getKey( context )
	log( key )

	updateTo := false
	json.NewDecoder( context.Request().Body ).Decode( &updateTo )

	if updateTo {
		log( "refreshing timestamp" )
		lastInteractionMutex.Lock()
		lastInteraction[key] = getTimeStamp()
		lastInteractionMutex.Unlock()
	}

	return context.String( http.StatusOK, "ok" )
}


func getKey(context echo.Context) (string) {
	var address = context.Param( "address" )
	return address
}


func log( text string ) {
	fmt.Println( getTimeStamp() + " - " + currentFunctionName() + " - " + text )
}
func logError( text string ) {
	fmt.Println( getTimeStamp() + " - \033[1;31m" + currentFunctionName() + " - " + text + "\033[0m" )
}




func getOriginTimeStamp() string {
	loc, _ := time.LoadLocation( "America/New_York" )
	return time.Date( 1970, 1, 1, 0, 0, 0, 0, time.UTC ).In(loc).Format("2006-01-02 15:04:05.000000")
}

func getTimeStamp() string {
	loc, _ := time.LoadLocation( "America/New_York" )
	return time.Now().In(loc).Format("2006-01-02 15:04:05.000000")
}


func currentFunctionName() string {
	functionName := "anonymousFunction"
	pc, _, _, ok := runtime.Caller( 2 )
	if( ok ) {
	    fn := runtime.FuncForPC( pc )
	    if( fn!=nil ) {
	    	functionName = fn.Name()
	    }
	}

	functionName = strings.Replace( functionName, "main.", "", 1 )

    return functionName
}


func addToErrors( key string, errorMessage string ) {
	logError( currentFunctionName() + " - " + errorMessage )
	if _, ok := deviceErrors[key]; !ok {
		deviceErrors[key] = make(map[string]string)
	}
	deviceErrors[key][getTimeStamp()] = currentFunctionName() + " - " + errorMessage
}
func addToErrorsAndReturn(key string, errorMessage string, toReturn bool) bool {
	addToErrors(key, currentFunctionName() + " - " + errorMessage)

	return toReturn
}
func getErrors( context echo.Context ) error {
	var key = getKey( context )
	log( key )

	httpResponseCode := http.StatusOK
	errs, gotErrors := deviceErrors[key]
	if gotErrors {
		if len(errs) > 0 {
			httpResponseCode = http.StatusInternalServerError
			deviceErrorsMutex.Lock()
			deviceErrors[key] = make( map[string]string )
			deviceErrorsMutex.Unlock()
		} else {
			gotErrors = false
		}
	}

	return context.JSON( httpResponseCode, errs )
}
