/**
 * struct containing a custom command
 * The parameters are optional but, if present, is expected to be a marshalled JSON object
 */
struct Command {
	1: string name,
	2: optional binary parameters,
}

/**
 * exception containing an error message
 * exceptions convert into a GO error.
 * By default, all service methods return an error but exceptions are handy if you require an error type that is
 * visible in both the client and server side code.
 */
exception SimpleError {
	1: string message
}

service SimpleService {
	/**
	* customCommand checks how custom commands can be passed to the server
	*/
	string customCommand(1:string id, 2:Command cmd) throws (1:SimpleError err),

	/**
     * snooze sleeps for the supplied number of seconds
     */
    void snooze(1:string id, 2:i64 secs),
}
