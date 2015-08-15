# Features
For this project to be viable over a standard CI platform it needs to have the following base features built in.

- [] Swappable sources (github, gitlab, bitbucket, filesystem). V1 will probably only have Github, but it should be possible to swap out the underlying source. 
	#Note, no matter what, the source must be a git repository of some sort. 
- [] Everything should be event/hook based, and users should be able to run code before and after built in functions. 
- [] All built in functions should be wrapped in conditionals, which can be turned off via the config file or environmental variables
- [] There should be a Dry Run mode of some sort, allowing the user to run the CI, without actually doing the final merging/
- [] Must support atleast chef and node packages to start, inheriting from a empty general spec
- [] Must automatically bump the package version by a minor value. 
- [] Must validate the package version does not conflict with an existing package. 
- [] Allow the commands being run in the base 