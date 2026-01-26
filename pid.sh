#!/bin/bash
for ((i=0; i < 100; i++))
do
	echo $i
	echo "Ti pidor!" >> /home/user/Desktop/pidor{$i}.txt
done
