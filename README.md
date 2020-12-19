# Kwanjai

### V2.0.0
[] Change database to PostgreSQL  
[] Add Redis for Caching  
[] Change frontend to React  

### API Endpoints
The root path is `/api`.

|endpoint|method|data requied|authenticated require|response|
|-|-|-|-|-|
|`/authentication/login`|POST|`{"id": string, "password": string}` Both email and username can be used for loggin in.|no|`{"message": string, "token": {"access_token": string, "refresh_token": string}}`|
|`/authentication/register`|POST|`{"username": string, "email": string, "password": string}`|no|`{"message": string, "token": {"access_token": string, "refresh_token": string}}`|
|`/authentication/logout`|POST|`{"refresh_token": string}` Refresh token is needed to be removed from database.|yes|`{"message": string}`|
|`/authentication/verify_email/:ID`|POST|`{"key": string}`|no|`{"message": string}`|
|`/authentication/resend_verification_email`|POST|`{"email": string}`|no|`{"message": string}`|
|`/authentication/token/refresh`|POST|`{"refresh_token": string}`|no|`{"token": {"access_token": string}}`|
|`/authentication/token/verify`|GET||yes|The endpoint is use for token lifetime. If token is not expired, it returns status 200.|
|`/user/all`|GET||yes|{`"message": string, "usernames": []string}`|
|`/user/my_profile`|GET||yes|{`{"message": string, "profile": {"username": string, "email": string, "fristname": string, "lastname": string, "password": string, "is_superuser": bool, "is_verified": bool, "is_active": bool, "joined_date": string, "profile_picture": string, "plan": string, "projects": int, "date_of_subscription": int}}` Raw password is neither going to be revealed nor even store in database. The password field in this object is either 'password_is_created' or blank. It is for indicating if password is set or not. The profile_picture field is the url of user profile pictrue.|
|`/user/update_password`|POST|`{"old_password": string, "new_password1": string, "new_password2": string}`|yes|`{"message": string}`|
|`/user/update_profile`|PATCH|`{"firstname": string, "lastname": string}` Both firstname and lastname are optional. If they are empty, nothing is going to be updated.|yes|`{"message": "Profile updated."}`|
|`/user/profile_picture`|PATCH|`"file"` (Form)|yes|`{"message": "Uploaded."}`|
|`/user/pay`|POST|`{"token": string, "price": int}`|yes|`{"message": "Subscribed successfully."}`|
|`/user/unsubscribe`|POST||yes|`{"message": "Unsubscribed successfully."}`|
|`/project/all`|GET||yes|`{"projects": []{"id": string, "user": string, "name": string, "members": []string, "added_date": string}}`|
|`/project/new`|POST|`{"name": string}`|yes|`{"message": string, "project": {"id": string, "user": string, "name": string, "members": []string, "added_date": string}}`|
|`/project/find`|POST|`{"id": string}`|yes|`{"message": string, "project": {"id": string, "user": string, "name": string, "members": []string, "added_date": string}}`|
|`/project/update`|PATCH|`{"id": string, "name": string, "members": []string}` Name must not be empty. Members must include project owner.|yes|`{"message": string, "project": {"id": string, "user": string, "name": string, "members": []string, "added_date": string}}`|
|`/project/delete`|DELETE|`{"id": string}`|yes|`{"message": string}`|
|`/board/all`|POST|`{"project: string"}`|yes|`{"boards": []{"id": string, "user": string, "name": string, "project": string, "position": int}}`|
|`/board/new`|POST|`{"name": string, "project": string}`|yes|`{"message": string, "board": {"id": string, "user": string, "name": string, "project": string, "position": int}}`|
|`/board/find`|POST|`{"id": string}`|yes|`{"message": string, "project": {"id": string, "user": string, "name": string, "project": string, "position": int}}`|
|`/board/update`|PATCH|`{"id": string, "name": string, "position": int}` New position must be in the range oldPosition-1 >= newPosition >= oldPosition+1. Name must not be empty. |yes|`{"message": string, "project": {"id": string, "user": string, "name": string, "project": string, "position": int}}`|
|`/board/delete`|DELETE|`{"id": string}`|yes|`{"message": string}`|
|`/post/all`|POST|`{"project": string}`|yes|`{"posts": []{"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": []{"uuid": string, "user": string, "body": string, "added_date": string, "last_modified": string}, "people": []string, "added_date": string, "last_modified": string, "due_date": string}}`|
|`/post/new`|POST|`{"board": string, "title": string, "content": string, "due_date": string}`|yes|`{"message": string, "post": {"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": [], "people": [], "added_date": string, "last_modified": string, "due_date": string}}` comments and people are empty array when post is created.|
|`/post/update`|POST|`{"id": string}`|yes|`{"message": string, "post": {"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": []{"uuid": string, "user": string, "body": string, "added_date": string, "last_modified": string}, "people": []string, "added_date": string, "last_modified": string, "due_date": string}}`|
|`/post/update`|PATCH|`{"id": string, "board": string, "title": string, "content": string, "due_date": string, "completed": bool, "urgent": bool, "people": []string, "due_date": string}` id is required. Other fileds are optional.|yes|`{"message": string, "post": {"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": []{"uuid": string, "user": string, "body": string, "added_date": string, "last_modified": string}, "people": []string, "added_date": string, "last_modified": string, "due_date": string}}`|
|`/post/delete`|DELETE|`{"id": string}`|yes|`{"message": string}`|
|`/post/comment/new`|POST|`{"id": string, "comments": [{"body": string}]}`|yes|`{"message": string, "post": {"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": []{"uuid": string, "user": string, "body": string, "added_date": string, "last_modified": string}, "people": []string, "added_date": string, "last_modified": string, "due_date": string}}`|
|`/post/comment/update`|PATCH|`"id": string, "comments": [{"uuid": string}]}`|yes|`{"message": string, "post": {"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": []{"uuid": string, "user": string, "body": string, "added_date": string, "last_modified": string}, "people": []string, "added_date": string, "last_modified": string, "due_date": string}}`|
|`/post/comment/delete`|DELETE|`"id": string, "comments": [{"uuid": string}]}`|yes|`{"message": string, "post": {"id": string, "board": string, "project": string, "user": string, "title": string, "content": string, "completed": bool, "urgent": bool, "comments": []{"uuid": string, "user": string, "body": string, "added_date": string, "last_modified": string}, "people": []string, "added_date": string, "last_modified": string, "due_date": string}}`|

Kwanjai is named after Thai music band KhanaKwanjai. You can enjoy their music [here](https://www.youtube.com/c/%E0%B8%84%E0%B8%93%E0%B8%B0%E0%B8%82%E0%B8%A7%E0%B8%B1%E0%B8%8D%E0%B9%83%E0%B8%88 "KhanaKwanjai youtube channel").