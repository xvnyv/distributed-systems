import json
import os

dir_name = os.path.dirname(os.path.realpath(__file__)) + "/results_report"
all_req_types = ["read", "write", "read-write"]
all_req_rates = [10, 100, 500]
all_nodes = [5, 10, 20]
data = {}  # {node: {req_type: {req_rate: {}}}}
for n in all_nodes:
    data[n] = {}
    for rt in all_req_types:
        data[n][rt] = {}
        for rr in all_req_rates:
            data[n][rt][rr] = {}

for fname in os.listdir(dir_name):
    lst_fname = fname.lstrip("results-N").rstrip(".json").split("_")
    nodes, req_type, req_rate = lst_fname
    req_rate, test_num = req_rate.lstrip("R").split("-")
    req_type = req_type.lower()

    # print(
    #     f"Nodes: {nodes}\nRequest: {req_type}\nRate: {req_rate}\nTest Number: {test_num}"
    # )

    with open(f"{dir_name}/{fname}", "r") as f:
        cur_data = json.load(f)
        data[int(nodes) + 1][req_type][int(req_rate)][int(test_num)] = cur_data
        # data format: {"SuccessRate":1,"PercentageBelow5ms":98.5,"AvgLatency":3.050125}

print(
    "%10s%15s%20s%10s%10s%10s%15s%15s%15s%15s%20s"
    % (
        "# Nodes",
        "Req Type",
        "Req Rate (/s)",
        "% <5ms 1",
        "% <5ms 2",
        "% <5ms 3",
        "% <5ms Avg",
        "Avg Latency 1",
        "Avg Latency 2",
        "Avg Latency 3",
        "Avg Latency Avg",
    )
)
csv = ""
for n, v in data.items():
    for rt, v2 in v.items():
        for rr, v3 in v2.items():
            total_latency = 0
            total_percentage = 0
            for i, d in v3.items():
                total_latency += d["AvgLatency"]
                total_percentage += d["PercentageBelow5ms"]
            avg_latency = round(total_latency / 3, 3)
            avg_percentage = round(total_percentage / 3, 2)
            print(
                "%10d%15s%20d%10s%10s%10s%15s%15s%15s%15s%20s"
                % (
                    n,
                    rt,
                    rr,
                    round(v3[1]["PercentageBelow5ms"], 2),
                    round(v3[2]["PercentageBelow5ms"], 2),
                    round(v3[3]["PercentageBelow5ms"], 2),
                    avg_percentage,
                    round(v3[1]["AvgLatency"], 3),
                    round(v3[2]["AvgLatency"], 3),
                    round(v3[3]["AvgLatency"], 3),
                    avg_latency,
                )
            )
            csv += f"{n},{rt},{rr},{round(v3[1]['PercentageBelow5ms'], 2)},{round(v3[2]['PercentageBelow5ms'], 2)},{round(v3[3]['PercentageBelow5ms'], 2)},{avg_percentage},{round(v3[1]['AvgLatency'], 3)},{round(v3[2]['AvgLatency'], 3)},{round(v3[3]['AvgLatency'], 3)},{avg_latency}\n"

with open("results.csv", "w") as f:
    f.write(csv)
