## hn-scan

This is a simple command-line tool written in Go which scans the current top HackerNews stories and filters the titles by some keywords which are provided as arguments.

Example usage:
```
> go run main.go -n 100 -term go -term microservice
Retrieved the following stories from the top 100 matching terms ["go" "microservice"]:

Show HN: A HN clone writen in Go (https://remoterenters.com/)
```

The application retrieves the current top 100 stories on HackerNews and filters for those that have "go" or "microservice" in the title.

## A note on the overall architecture

The project is implemented using a hexagonal architecture. This may seem like overkill for a project that is both small and has its scope clearly defined and limited ahead of time. But my main goal here is to learn to develop in Go, so I wanted to architect this like a serious project.

Plus, in my opinion, following the hexagonal architecture actually incurs a very small overhead even for the smallest of projects once you have the blueprint down. And it is almost always worth it for the advantages it gives: easy testability plus the fact that the architecture guides the steps to take to a clean design (not to mention the other advantages that are relevant in bigger projects and where teamwork is involved).

That said, the project isn't "overengineered" in other aspects of it. Structure aside, my goal is to make the sensible implementation decisions along the way.

So the way the code is structured is like this: there is a `core` subdirectory, which contains the business logic code (simple as it may be in this case). The core is meant to not have any dependency on infrastructure code and to be easily testable. The infrastructure layer depends on the core, not the other way around. The infrastructure layer is everything around the core.

The `core` defines inbound ports, in the form of the `HNService` struct, which has a high level method for retrieving the stories given the number of stories and keyword terms to filter on. This struct is called by the infrastructure layer directly in response to (command-line, in this case) invocations from the outside.

And the core also defines outbound ports in the form of the `TopStoriesRepo` interface inside `core/ports`. This interface is meant to be implemented by an adapter defined in the infrastructure layer: in this case, the `TopStoriesRepo` struct defined in `adapters/stories_repo.go`.

## Implementation notes

These are some notes on design decisions that are exemplified in this project and/or which are good keep in mind. These notes are mainly here for my own reference, as my goal here is to learn Go, so they include general Go guidelines:

- We don't want to create abstractions (interfaces) unless we know we need them (not anticipate that they may be needed). And interfaces should almost always be defined on the client side, of course. But sometimes, being able to mock for testing purposes is reason enough to define an interface. That's what I did in `cmd/cmd_app.go` by defining the `hNService` interface. I could have just accepted the struct defined in `core/hnservice.go`, but that would have broken my test isolation. (And it's the same in `adapters/stories_repo.go` with the `hackerNewsClient` interface.) So I accepted an interface and returned a struct as is usually the way to go. Good examples of "abstractions being discovered and not created".

- There are a few interfaces declared in the code for various reasons (`HackerNewsClient`, `HNService` and `TopStoriesRepo`). Variables of these interface types are assigned either structs or pointers to structs as needed. Note that when you assign a pointer-to-struct to an interface, you can call both value methods and pointer-receiver methods. But when you assign a value struct to an interface, the struct must implement all the interface methods with a value receiver. This is why interfaces almost always store pointers to structs (and the decision on the type of receiver is up to the struct methods).
This is not special to interfaces, though, interfaces are just a special case. "In Go, a method with a value receiver can be called on both a struct value and a pointer to the struct. When a method is called on a pointer to a struct, Go automatically dereferences the pointer and calls the method on the underlying struct value (a copy is still made, because the receiver is value and thus, the original is unmodified). This is known as pointer indirection. However, a value struct type cannot call methods with a pointer receiver because the method may modify the struct value, and passing the struct by value would only modify a copy of the struct."

- I don't have any actual pointers to interfaces in the code, this is almost never needed, except for some very specialized use cases.

- For the `HNServiceRequest` (part of the core and the only parameter to the main `HNService` method), I went with what is essentially a "custom config struct pattern". (The `HNServiceRequest` struct being the "config" struct). Because I wanted to encapsulate validation logic, I made the struct fields unexported and defined a function that creates such a request from given parameters. What this gives us over simply having the `HNService.GetTopStories` method accept the params directly is a bit of separation of concerns and a more stable and usable API for our main service method. This approach can be extended in some ways, but it seemed like overkill in this case: For one, a Builder pattern could be used to construct our config/request. This would be particularly handy if we have a larger number of parameters and we need to distinguish the cases where arguments are passed vs. not passed (as opposed to just having the default value used); but this isn't really the case here. Another idiomatic approach often used is the **functional options pattern**. This would be a particularly API friendly pattern that is often used in Go. But it does imply that we have to move the default value handling inside our service method, so we lose some of that encapsulation I was talking about. So it didn't seem worth it in this case, but it's good to be aware of it.

- Speaking of not needing to distinguish the case when a variable was explicitly initialized or not: in Go, it's usually ok to use nil slices to avoid unnecessary allocations. This is because nil slices and empty slices can be used interchangeably when calling `append` and for iteration. So no need to make defensive allocations when returning slices just for the sake of not returning nil. So might as well avoid an unnecessary allocation. That's why in `NewHNServiceRequest`, there is no handling of the case of terms being nil. Either way, it is used as provided. (Note that this doesn't mean there aren't also some cases where libraries will make a difference between the nil vs empty cases - like when rendering json).