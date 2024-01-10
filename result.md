Problem 3
The result should not always be zero, because with preemtive scheduling swapping can occur at any time.
Here we can see that the work of one thread can be overwritten by another. 
This is called a race condition, i.e. the result depends on ordering in time.
The first runs resulted in;
388644
-264823
-836222
46800
...

Problem 4
C: Mutex. because only the thread that locks can unlock. We only have two threads so we do not need higher numbers than 1, thus we can use binary semaphore with mutex.
This works. get 0 every time