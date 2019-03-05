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
	* ping the service
	*/
	void ping(),

	/**
	* getString returns a string containing the supplied id
	*/
	string getString(1:string id),

	/**
	* runCustomCommand checks how custom commands can be passed to the server
	*/
	string runCustomCommand(1:string id, 2:Command cmd) throws (1:SimpleError err),
}
