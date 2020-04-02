import lyx

lines = lyx.io.read_all_lines("input.txt")
res = [len(x.split(" ")) for x in lines]
num = sum(res)
print('\n\nnum:')
print(num)
