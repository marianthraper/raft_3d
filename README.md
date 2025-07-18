# Raft_3D

Backend API for managing a 3D printing system.

The Raft Consensus algorithm is used to manage persistence and consistency instead of a database.

The project is implemented in Go using the hashicorp/raft library.

It project entails:
1. Leader election 
2. Event driven architecture
3. Eventual Consistency
4. Failover management 
5. Maintaining event log and snapshotting for fault tolerance

The different objects managed include - printers, print_jobs and filaments

A finite state machine (FSM) is used to handle various processes for these objects including creating, listing and updating while ensuring consistency. 

