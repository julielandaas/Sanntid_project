Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> In concurrency we switch between threads fast, so it looks like its happening in paralell. In parallellalism it physically runs in multiple processes at the same time.

What is the difference between a *race condition* and a *data race*? 
> race condition: the result is dependent of the order the tasks are executed. A data race is a type of race condition when at least one is trying to write and at least one is trying to read from the same memory location, and there is no locks  to prevent this.
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> Decides which threads are happening next. It holds lists of runnable and blocked threads and chooses one of the runnable threads to execute next.


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> We use multiple threads to enhance performance and use the CPU more efficiently when we want multiple different things to happen at the same time. It also improves te code quality and makes the code easier to read.

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
>The goal with using Fibers is to divide the thread into several tasks - independent tasks do their own work, as opposed to work being passed around as functions. This makes it "look" like blocking
OS threads
fibers: programmeringspråk - høyere opp, med egen scheduler

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> A bit moore work in the begiinning, but makes it a lot easier to change and modify the code

What do you think is best - *shared variables* or *message passing*?
> Using go we prefer message passing


