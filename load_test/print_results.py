import json
import os

SCRIPT_DIR = os.path.dirname(os.path.realpath(__file__))

all_req_types = ["read", "write", "read-write"]
all_req_rates = [10, 100, 500]
all_nodes = [5, 10, 20]


def create_data_map():
    data = {}  # {node: {req_type: {req_rate: {}}}}
    for n in all_nodes:
        data[n] = {}
        for rt in all_req_types:
            data[n][rt] = {}
            for rr in all_req_rates:
                data[n][rt][rr] = {}

    return data


def get_test_attr(fname, fext=".json"):
    lst_fname = fname.lstrip("results-N").rstrip(fext).split("_")
    nodes, req_type, req_rate = lst_fname
    req_rate, test_num = req_rate.lstrip("R").split("-")
    req_type = req_type.lower()
    return int(nodes) + 1, req_type, int(req_rate), int(test_num)  # int, int, str, int


def output_table_and_csv():
    dir_name = SCRIPT_DIR + "/results_report"
    data = create_data_map()
    for fname in os.listdir(dir_name):
        nodes, req_type, req_rate, test_num = get_test_attr(fname)

        with open(f"{dir_name}/{fname}", "r") as f:
            cur_data = json.load(f)
            data[nodes][req_type][req_rate][test_num] = cur_data
            # data format: {"SuccessRate":1,"PercentageBelow5ms":98.5,"AvgLatency":3.050125}

    print(
        "%10s%15s%20s%20s%20s%20s%25s%15s%15s%15s%20s"
        % (
            "# Nodes",
            "Req Type",
            "Req Rate (/s)",
            "99th Percentile 1",
            "99th Percentile 2",
            "99th Percentile 3",
            "99th Percentile Avg",
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
                total_percentile = 0
                for _, d in v3.items():
                    total_latency += d["AvgLatency"]
                    total_percentile += d["99thPercentile"]
                avg_latency = round(total_latency / 3, 3)
                avg_percentile = round(total_percentile / 3, 3)
                print(
                    "%10d%15s%20d%20s%20s%20s%25s%15s%15s%15s%20s"
                    % (
                        n,
                        rt,
                        rr,
                        round(v3[1]["99thPercentile"], 3),
                        round(v3[2]["99thPercentile"], 3),
                        round(v3[3]["99thPercentile"], 3),
                        avg_percentile,
                        round(v3[1]["AvgLatency"], 3),
                        round(v3[2]["AvgLatency"], 3),
                        round(v3[3]["AvgLatency"], 3),
                        avg_latency,
                    )
                )
                csv += f"{n},{rt},{rr},{round(v3[1]['99thPercentile'], 3)},{round(v3[2]['99thPercentile'], 3)},{round(v3[3]['99thPercentile'], 3)},{avg_percentile},{round(v3[1]['AvgLatency'], 3)},{round(v3[2]['AvgLatency'], 3)},{round(v3[3]['AvgLatency'], 3)},{avg_latency}\n"

    with open("results.csv", "w") as f:
        f.write(csv)


def get_99_percentile():
    dir_name = SCRIPT_DIR + "/results_json"
    for fname in os.listdir(dir_name):
        with open(f"{dir_name}/{fname}", "r") as f:
            cur_data = json.load(f)
            percentiles = []
            for m in cur_data:
                percentiles.append(m["latencies"]["99th"])
            percentile = max(percentiles) / 1_000_000

            with open(f"{SCRIPT_DIR}/results_report/{fname}", "r+") as f_report:
                report_data = json.load(f_report)
                report_data["99thPercentile"] = percentile
                f_report.seek(0)
                json.dump(report_data, f_report)
                f_report.truncate()


output_table_and_csv()
