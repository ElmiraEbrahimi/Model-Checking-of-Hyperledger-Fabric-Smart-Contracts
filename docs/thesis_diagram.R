#/////////////////////////////////////////////////////////asset_delively
library(ggplot2)
totalTransaction <- c(100,200,400,600,800,1000,2000)
speedUp<- c(94.2,87.6,89.9,80.48,116.3,111.3,170.5)
dat <- data.frame( totalTransaction, speedUp)
ggplot(dat, aes(x =  totalTransaction, y = speedUp) , label= speedUp) + 
  geom_point()+geom_line(color="blue") + geom_text(aes(label=speedUp),hjust=0, vjust=0)+
  scale_color_brewer(palette="Paired")+
  theme_classic()+labs(y= "Speed Up", x = "Transaction Total Number (Tx)")


#////////////////////////////////////////////////////////////

library(ggplot2)
totalTransaction <- c(100,200,400,600,800,1000,2000)
contract_without_bugs<- c(50766,109607,210833,343433,439552,547439,1128088)*0.001
contract_with_bugs<- c(4785650,9611051,18953457,25237918,51124189,60931083,192851356)*0.001

df <- data.frame(contract_without_bugs,contract_with_bugs)
df <- stack(df)
df$totalTransaction <- totalTransaction
colnames(df) <- c("sumExecutionTime", "results", "totalTransaction")
ggplot(df, aes(x=totalTransaction, y=sumExecutionTime, colour=results)) +
  geom_point()+ geom_line() +theme_classic()+labs(y= "Total Execution Time (s) ", x = "Transaction Total Number (Tx)")


#/////////////////////////////for server 
library(ggplot2)
totalTransaction <- c(100,200,400,600,800,1000,2000)
speedUp<- c(81.4,70.8,67.1,64.9,92.4,130.4,132.1)
dat <- data.frame( totalTransaction, speedUp)
ggplot(dat, aes(x =  totalTransaction, y = speedUp) , label= speedUp) + 
  geom_point()+geom_line(color="blue") + geom_text(aes(label=speedUp),hjust=0, vjust=0)
+theme_classic()+labs(y= "Speed Up", x = "Transaction Total Number (Tx)")




#.....................................

library(ggplot2)
totalTransaction <- c(100,200,400,600,800,1000,2000)
sumValidInServer<- c(63980,144963,306914,468472, 622397,852162,1824378)*0.001
sumInvalidInServer<- c(5211513,10275964,20601641,30419344,57542311,111128437,233624777)*0.001

df <- data.frame(sumValidInServer,sumInvalidInServer)
df <- stack(df)
df$totalTransaction <- totalTransaction
colnames(df) <- c("sumExecutionTime", "results", "totalTransaction")
ggplot(df, aes(x=totalTransaction, y=sumExecutionTime, colour=results)) +
  geom_line() +theme_classic()


#.....................................

library(ggplot2)
totalTransaction <- c(100,200,400,600,800,1000,2000)
speedUp_without_server<- c(94.2,87.6,89.9,80.48,116.3,111.3,170.5)
speedUp_with_server<- c(81.4,70.8,67.1,64.9,92.4,130.4,132.1)
df <- data.frame(speedUp_without_server,speedUp_with_server)
df <- stack(df)
df$totalTransaction <- totalTransaction
colnames(df) <- c("speedUp", "results", "totalTransaction")
ggplot(df, aes(x=totalTransaction, y=speedUp, colour=results)) +
  geom_line() +theme_classic()+  geom_point()+labs(y= "Speed Up", x = "Transaction Total Number (Tx)")



#.....................................

library(ggplot2)
totalTransaction <- c(100,200,400,600,800,1000,2000)
speedUp<- c(81.4,70.8,69.1,64.9,92.4,120.4,132.1)
dat <- data.frame( totalTransaction, speedUp)
ggplot(dat, aes(x =  totalTransaction, y = speedUp) , label= speedUp) + 
  geom_point()+geom_line(color="blue") + geom_text(aes(label=speedUp),hjust=0, vjust=0)+theme_classic()+labs(y= "Speed Up", x = "Transaction Total Number (Tx)")





