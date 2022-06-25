If a go run command is executed from the directory the go.work
file is located in then it will use that go.work file

If you are trying to go run/go build from the submodule
where no go.work file exists then it will not pick the go.work file
up.

This can be alleviated by setting the GOWORK file to the absolute path
of the parent directory.

I am trying to figure out what the best practice could be for a big 
multi module repository. Allegedly it is suggested to not commit the 
go.work file to source control but I am not sure this makes sense in the 
otel use case.

We already commit all the replace statements so what would be the difference 
in committing a go.work file?

I should finish reading all the docs to see why it is suggested to not 
commit to source control. I also wonder if crosslink can be expanded in a way 
which can temporarily create workspace files based on the best practice.

./crosslink work create
   creates go.work file
   sets the env var

./crosslink work destroy
   Deletes work file
   unsets env var

Is there a point to creating a gowork file in every subdirectory rather
than setting the GOWORK env var? I don't think this makes much sense and defeats
the point of the of the work file. 

As a starting point the crosslink tool could be extended to inlcude the work create
command. The `work create` command would create a `go.work` file in the current directory.
Dependencies would be automatically added so that a user could quickly get a go.work file 
up and running.
