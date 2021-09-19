import jsonlines
from datetime import datetime
import matplotlib
import structlog
import matplotlib.pyplot as plt
import matplotlib.dates as mdate

log = structlog.get_logger()

server_stats = {}
with jsonlines.open("servers.jl") as reader:
    for line in reader:
        if line["invite_id"] in server_stats:
            server_stats[line["invite_id"]].append(line)
        else:
            server_stats[line["invite_id"]] = [line]

fig, axs = plt.subplots(len(server_stats), figsize=(15, 30))

# Customize
SMALL_SIZE = 8
MEDIUM_SIZE = 10
BIGGER_SIZE = 12

plt.rc('font', size=SMALL_SIZE)          # controls default text sizes
plt.rc('axes', titlesize=SMALL_SIZE)     # fontsize of the axes title
plt.rc('axes', labelsize=MEDIUM_SIZE)    # fontsize of the x and y labels
plt.rc('xtick', labelsize=SMALL_SIZE)    # fontsize of the tick labels
plt.rc('ytick', labelsize=SMALL_SIZE)    # fontsize of the tick labels
plt.rc('legend', fontsize=SMALL_SIZE)    # legend fontsize
plt.rc('figure', titlesize=BIGGER_SIZE)  # fontsize of the figure title

i = 0
for server_id in server_stats:
    if len(server_stats[server_id]) > 1:
        # First sort the instances of a single server by time
        date_format = '%Y-%m-%d %H:%M:%S'
        single_server_instances = sorted(server_stats[server_id], key=lambda x: datetime.strptime(
            x["scrape_date"], date_format))
        # Identify the first and last instances of the server scrape
        first_instance = single_server_instances[0]
        last_instance = single_server_instances[len(server_stats[server_id])-1]
        server_name = first_instance["server_name"]
        # Log the growth from the first scrape to the last
        growth = last_instance["member_count"] - \
            first_instance["member_count"]
        first_date = datetime.strptime(
            first_instance["scrape_date"], '%Y-%m-%d %H:%M:%S')
        last_date = datetime.strptime(
            last_instance["scrape_date"], '%Y-%m-%d %H:%M:%S')
        time_between = last_date - first_date
        time_between_seconds = time_between.seconds
        hourly_growth = round(growth/(time_between.seconds/3600), 1)
        log.msg("Server growth", server_name=server_name,
                growth=growth, time_between_seconds=time_between_seconds, hourly_growth=hourly_growth)

        # Plot
        list_of_dates = []
        counts = []
        for instance in single_server_instances:
            list_of_dates.append(instance["scrape_date"])
            counts.append(instance["member_count"] -
                          first_instance["member_count"])

        dates = matplotlib.dates.date2num(list_of_dates)
        axs[i].set_title(server_name)
        axs[i].plot_date(dates, counts, linestyle="solid")
        date_formatter = mdate.DateFormatter(date_format)
        axs[i].xaxis.set_major_formatter(date_formatter)

    else:
        print("There is only one instance of " +
              server_stats[server_id][0]["server_name"] + ": " + str(server_stats[server_id][0]["member_count"]) + " members")
    i += 1

fig.tight_layout()
fig.autofmt_xdate()
plt.savefig("servers.png", bbox_inches='tight', dpi=1000.0)
plt.show()
