paths = "_source.kubernetes.cluster_name"
JSON = {

        "_source":{
            "kubernetes":{
                "cluster_name":"name"
            }
        }
    
}

def find(element, JSON):     
    paths = element.split(".")
    data = JSON
    try:
        for i in range(0,len(paths)):
            data = data[paths[i]]
            print(data)
    except:
        return 'Empty'
    return data


# print(find(paths, JSON))


find(paths, JSON)