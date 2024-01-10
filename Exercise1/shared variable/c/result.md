The result should not always be zero, because with preemtive scheduling swapping can occur at any time.
Here we can see that the work of one thread can be overwritten by another. 
This is called a race condition, i.e. the result depends on ordering in time.
The first runs resulted in;
388644
-264823
-836222
46800
...