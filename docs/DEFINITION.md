# Nexema File Structure

A simple Nexema file may look like this:

```go
include "v1/a.nex";
include "v1/foo/bar.nex" as foo_bar;
include "v1/requests.nex" as requests;

type Entity base {
    id string
}

type User extends Entity {
    first_name      string
    phone_number    varchar(11)?
    tags            list(string)
    preferences     map(string, bool)
    account_type    AccountType

    __defaults {
        tags = ["plain_user"]
        preferences = {
            "cats": true,
            "dogs": false
        }
    }
}

type AccountType enum {
    // An unknown account type
    unknown

    // An admin account type
    admin

    // A customer account type
    customer
}

#obsolete = true
type Account union {
    1 foo_bar.Admin     admin
    2 foo_bar.Customer  customer
}

service RegistrationService {

    User register(requests.RegisterRequest);

    #http_method = GET
    User? getUser(requests.GetUserRequest);

    downstream User onUpdate(requests.ListenUserRequest);

    upstream void sendMessage(request.NewMessageRequest);

    bidirectional Message messages(request.MessagesRequest);
}
```
