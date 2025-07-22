In our project we want to make some properties to public while some of them are private how can we achieve this? For instance, the transaction between supplier and car factory can be private since we want to keep their trading as a secret while customer can validate without reveling their sensitive data. Should we use an authority in this case.
-  That really depends on the model that we want to use. If we use a Blockchain Database we could manage access trough some identifier from the requester.  If we do something different  like e.g. hyperledger fabric things can be managed through channels. Mehmet what have you worked with so far? Do you maybe have an idea for a infrastructure?


As far as I know one of the most critical part of car’s considering environmental approach is their batteries. There was a battery passport example I guess it is closely related to SSI. In this case should we consider create a verifiable identity(SSI) for these cars?
- I guess that would make sense yes 

In our project we want to focus on supply chain of car production and we specify our approach to environmental friendliness. Should we broaden our approach on this project.
- I don't think that we have to focus on cars necessarily. But sustainability does seem to be a requirement. 

In our project some of the stake holders such as car companies, suppliers need to join network with specified credentials or not every one that authenticated by us can join the network with these specified roles. However, users such customers can be any one in our project. What kind of blockchain should we used in that sense? Should not we give customers join the blockchain network?
- Also really depends on the structure. If we use a BCDB we could have a front-end that automatically gives users access to the information that they are supposed to see. If it's hyperledger fabric then there could be a channel for customers
