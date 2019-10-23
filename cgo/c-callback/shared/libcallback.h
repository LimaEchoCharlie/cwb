#ifndef __CALLBACK_H_
#define __CALLBACK_H_

typedef enum Result {Failure = -1, Success = 0} Result;
typedef Result (*callback)(char*);

#endif
