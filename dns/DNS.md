# DNS idk
DNS is comprised of:
```
+-------------------+
|   Header          |
+-------------------+
|   Question        | \# The question
+-------------------+
|   Answer          | \# Resource Records answering the question
+-------------------+
|   Authority       | \#RRs pointing towards an autority
+-------------------+
|   Additional      | \#RRs holding additional info
+-------------------+
```

## DNS Header 
```
0  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      ID                       |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|QR|   OPCODE  |AA|TC|RD|RA|   Z    |   RCODE   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    QDCOUNT                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ANCOUNT                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    NSCOUNT                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ARCOUNT                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## DNS Question
```
0  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                     QNAME                     / \# See below
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     QTYPE                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     QCLASS                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```
## DNS Recourse Record ( Answer / Authority /Additional ) 
```
0  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                     NAME                      /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     TYPE                      |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     CLASS                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     TTL                       \|
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                   RDLENGTH                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                    RDATA                      /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

# Header explainers
### QR 
A one-bit field, 0 for query, 1 for response

### OPCODE
A four-bit field that specifies kind of query
Set by orignator of query and put into response.
| Opcode  | Name                             | Reference |
|---------|----------------------------------|-----------|
| 0       | Query                            | \[RFC1035] |
| 1       | IQuery (Inverse Query, OBSOLETE) | \[RFC3425] |
| 2       | Status                           | \[RFC1035] |
| 3       | Unassigned                       |           |
| 4       | Notify	\[RFC1996]                |           |
| 5       | Update                           | \[RFC2136] |
| 6       | DNS Stateful Operations (DSO)    | \[RFC8490] |
| 7 to 15 | Unassigned                       |           |

### AA (Authoritive answer )

This bit is valid in responses, specifies that the responding nameserver is an authority for said domain name 

### TC 
Is the message truncated due to size

### RD 
Is the recursion desired ( set when making query ) 


### RA
Denotes if recursion is available in response from DNS server


### Z 
Reserved for future


### RCODe - Response code
4-fit field is set as part of responses

| Code | meaning         |
|------|-----------------|
| 0    | No error        |
| 1    | Format error    |
| 2    | Server failure  |
| 3    | Name Error      |
| 4    | Not Implemented |
| 5    | Refused         |


### QDCount 
16-bit, denotes number of questions in request


### ANCount
16-bit, denotes number of answers in response

### NSCount 
16-bit, denotes number of name server resouce records in authority records section....

### ARCount
16-bit, denotes number of RRs in additional records section






# Answer explainers

### QNAME

Contains domain names we wish to resolve.
Shown as a sequence of labels, each is one octet ( 8 bit ) length field followed by that number of octets.

The domain name terminates with the zero-length octet for the null label of the root.

E.g. example.com

Two sections, "example" and "com"

Labelled into ( ascii encoded - in theory case insensitive):
"example" -> 101 120 97 109 112 108 101
"com" -> 99 111 109

"example" has 7 chars, so will be proceded by int byte of len

"7 101 ....."
Put into binary... then put into qname


#### Compression QNAME 


"To cut down on duplication, a special technique called message compression is used. 
Instead of a DNS name encoded as above using the combination of labels and label-lengths, 
a two-byte subfield is used to represent a pointer to another location in the message where the name can be found. 
The first two bits of this subfield are set to one (the value “11” in binary), 
and the remaining 14 bits contain an offset that species where in the message the name can be found, 
counting the first byte of the message (the first byte of the ID field) as 0." 

[Source](http://www.tcpipguide.com/free/t_DNSNameNotationandMessageCompressionTechnique-2.htm)

### QType
2-BYTE value specifies type of query - see [Wikipedia](https://en.wikipedia.org/wiki/List_of_DNS_record_types)

### QClass
Mostly (1) for "IN" - "internet"

Used to have Chaosnet and Hesiod

# DNS RR ( Resource record)'s explainer

### NAME
The domain name this record is for (same encoding as QNAME)


### TYPE
What type of record this is (A, AAAA, MX etc.) same values as QTYPE

### CLASS
Same as QCLASS, almost always 1 (IN)

### TTL
Time To Live, how long (in seconds) you can cache this record
e.g. TTL=300 means "cache this for 5 minutes"

### RDLENGTH
The length in bytes of the RDATA field that follows

### RDATA
The actual answer data

This data length is variable e.g:

A record ( IPv4) - 4 bytes
AAAA Record ( IPv6) - 16 bytes
CNAME record - encoded same as QNAME 
MX record Two parts - 16-bit preference number ( priority) then domain name
NS Record - just a domain name


