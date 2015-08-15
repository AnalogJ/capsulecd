# capsulecd
Continuous Delivery scripts for automating package releases (npm, chef, ruby, python, crate)


#capsule.yml
capsume.yml file should have the following sections:

    # docker image which will be used to build and package this release. 
    # Image should have all required tools installed to validate, build, package and release 
    image: nodejs 
    
    # source should be where the package source is from. ie. github/bitbucket/gitlab (only github supported to start)
    source: 
      github:
        # additional github specific configuration which can be used to configure enterprise github connections
    
    # type specifies which capsule script is run against the code
    # can only be: npm, chef, ruby, python, crate, general, (more to come)
    type: npm
    
    # validate contains the configuration for the validation step. 
    validate:
      config:
        #flags is an array of options used to enable and disable steps in the default package validate script.
        flags: []
        #options which are used to 
      
      # the pre hook takes a multiline script which allows you to prepare the environment before the package validation script is run. This could include downloading additional test libraries, . This custom pre step will be run before the default package script is run. 
      pre:
      
      # the post hook takes a multiline ruby script which allows you to do any cleanup or post processing after the package validation script is run. This could include steps like running a custom test/validation suite, deleting test folders, commiting added files
      post:
      
    # the build hook generates the acutal package??
    build: 
      config:
        flags: []
      pre:
      post:
      
    # deploy the package to the package source (npm publish, gem push, etc)
    release:
      config:
        flags: []
      pre:
      post:
