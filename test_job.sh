#!/bin/bash
#
#SBATCH --partition=scrubpeak
#SBATCH --account=chpc

#SBATCH --job-name=test
#SBATCH --output=out.text
#SBATCH --nodelist=sp001
#SBATCH --ntasks=1
#SBATCH --time=10:00

echo "launching two cpu timeout 10s"
../stress-ng/stress-ng --cpu 2 --timeout 10s
echo "finished"
