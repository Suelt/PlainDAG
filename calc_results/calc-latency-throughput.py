#This file is used to calculate the latency and throughput
import os
import time
#ResultPath is the path where the results are stored
ResultsPath = "/home/kuokhane/shareWithPC/code/go/PlainDAG-DSN/ansible-remote/result"

def get_filelist(dir, Filelist):
    if os.path.isfile(dir):
        Filelist.append(dir)
    elif os.path.isdir(dir):
        for s in os.listdir(dir):
            newDir = os.path.join(dir, s)
            get_filelist(newDir, Filelist)
    return Filelist

if __name__ == '__main__':

    list = get_filelist(ResultsPath, [])
    nodesSum=len(list)
    print("nodeNumber",nodesSum)
    nodesSeq=0
    latencySum=0
    throughputSum=0
    startTime=0
    duration=0
    skipsum = 0
    Note = open('FinalResults.txt', mode='w')
    for nodeName in list:
        exeNum=0
        latency = 0
        nodesSeq=nodesSeq+1
        Note.write("Results of node"+ str(nodesSeq)+'\n')
        resultFile = open(nodeName, 'r', encoding='ISO-8859-1')
        for line in resultFile:
            if "latency" in line:
                if exeNum==0:
                    startTime=int(time.mktime(time.strptime(line[0:19],"%Y/%m/%d %H:%M:%S")))
                duration=int(time.mktime(time.strptime(line[0:19],"%Y/%m/%d %H:%M:%S")))-startTime
                exetime = int(line[line.index('cy:') + 3:])/1000
                latency=latency+exetime
            if "contained" in line:
                exeNum=exeNum+1
        resultFile.close()
        Note.write("Number of blocks executed: " + str(exeNum) + '\n')
        if exeNum==0:
            skipsum+=1
            continue
        AvgLatency=latency/exeNum
        latencySum=latencySum+AvgLatency
        duration=duration

        throughputSum=throughputSum+exeNum/duration
        Note.write("AvgLatency(s): " + str(AvgLatency) + '\n')
        Note.write("AvgThroughput(KRPS): " + str(exeNum/duration) + '\n')
    Note.write("==================================================\n")
    Note.write("TotalValidNodeSum: " + str(nodesSum-skipsum) + '\n')
    Note.write("==================================================\n")
    Note.write("AvgLatency(s): " + str(latencySum/(nodesSum-skipsum)) + '\n')
    Note.write("==================================================\n")
    Note.write("AvgThroughput(KRPS): " + str(throughputSum / (nodesSum-skipsum)) + '\n')
    Note.close()
